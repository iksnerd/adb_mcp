package tools

import (
	"context"
	"strings"

	"github.com/iksnerd/adb_mcp/internal/android"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Arguments ----

type listPackagesArgs struct {
	serialArg
	Filter string `json:"filter,omitempty" jsonschema:"Substring to filter package names."`
}

type installArgs struct {
	serialArg
	APKPath string `json:"apk_path" jsonschema:"Local filesystem path to the .apk to install."`
}

type packageArg struct {
	serialArg
	Package string `json:"package" jsonschema:"Application package name (e.g. com.example.app)."`
}

type permissionArgs struct {
	serialArg
	Package    string `json:"package" jsonschema:"Application package name."`
	Permission string `json:"permission" jsonschema:"Full permission name, e.g. android.permission.CAMERA."`
}

type openURLArgs struct {
	serialArg
	URL     string `json:"url" jsonschema:"URL or deep link to open (ACTION_VIEW)."`
	Package string `json:"package,omitempty" jsonschema:"Optional package to target the intent at."`
}

type lastCrashArgs struct {
	serialArg
	Package string `json:"package,omitempty" jsonschema:"Optional package name to filter to (e.g. com.example.app); omit for the most recent crash from any app."`
}

type pushArgs struct {
	serialArg
	LocalPath  string `json:"local_path" jsonschema:"Local file to copy onto the device."`
	DevicePath string `json:"device_path" jsonschema:"Destination path on the device, e.g. /sdcard/Download/x.json."`
}

type pullArgs struct {
	serialArg
	DevicePath string `json:"device_path" jsonschema:"File path on the device to copy off."`
	LocalPath  string `json:"local_path" jsonschema:"Local destination path."`
}

// ---- Handlers ----

func listPackages(ctx context.Context, in listPackagesArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	pkgs, err := android.ListPackages(ctx, serial, in.Filter)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return text("(no matching packages)"), nil
	}
	return text("%s", strings.Join(pkgs, "\n")), nil
}

func installApp(ctx context.Context, in installArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	out, err := android.InstallApp(ctx, serial, in.APKPath)
	if err != nil {
		return nil, err
	}
	return text("%s", out), nil
}

func launchApp(ctx context.Context, in packageArg) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	component, err := android.LaunchApp(ctx, serial, in.Package)
	if err != nil {
		return nil, err
	}
	if component != "" {
		return text("Launched %s (%s).", in.Package, component), nil
	}
	return text("Launched %s.", in.Package), nil
}

func stopApp(ctx context.Context, in packageArg) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.StopApp(ctx, serial, in.Package); err != nil {
		return nil, err
	}
	return text("Force-stopped %s.", in.Package), nil
}

func reloadApp(ctx context.Context, in packageArg) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.ReloadApp(ctx, serial, in.Package); err != nil {
		return nil, err
	}
	return text("Sent a reload broadcast to %s. Best-effort — if it didn't visibly reload (common on newer RN/Expo dev clients), use open_dev_menu then tap_on_text(\"Reload\") instead.", in.Package), nil
}

func openDevMenu(ctx context.Context, in serialArg) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.OpenDevMenu(ctx, serial); err != nil {
		return nil, err
	}
	return text("Opened the dev menu on %s. Use tap_on_text or describe_ui to pick an option (Reload, Debug JS Remotely, ...).", serial), nil
}

func uninstallApp(ctx context.Context, in packageArg) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	out, err := android.UninstallApp(ctx, serial, in.Package)
	if err != nil {
		return nil, err
	}
	return text("%s", out), nil
}

func clearAppData(ctx context.Context, in packageArg) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	out, err := android.ClearAppData(ctx, serial, in.Package)
	if err != nil {
		return nil, err
	}
	return text("Cleared data for %s (%s).", in.Package, strings.TrimSpace(out)), nil
}

func grantPermission(ctx context.Context, in permissionArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.GrantPermission(ctx, serial, in.Package, in.Permission); err != nil {
		return nil, err
	}
	return text("Granted %s to %s.", in.Permission, in.Package), nil
}

func revokePermission(ctx context.Context, in permissionArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.RevokePermission(ctx, serial, in.Package, in.Permission); err != nil {
		return nil, err
	}
	return text("Revoked %s from %s.", in.Permission, in.Package), nil
}

func openURL(ctx context.Context, in openURLArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	out, err := android.OpenURL(ctx, serial, in.URL, in.Package)
	if err != nil {
		return nil, err
	}
	return text("Opened %s.\n%s", in.URL, strings.TrimSpace(out)), nil
}

func lastCrash(ctx context.Context, in lastCrashArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	crash, found, err := android.LastCrash(ctx, serial, in.Package)
	if err != nil {
		return nil, err
	}
	if !found {
		if in.Package != "" {
			return text("(no recent crash for %s in the DropBox — it may not have crashed, or the entry rotated out)", in.Package), nil
		}
		return text("(no recent app crash in the DropBox)"), nil
	}
	return text("%s", crash), nil
}

func getAppDetails(ctx context.Context, in packageArg) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	d, err := android.GetAppDetails(ctx, serial, in.Package)
	if err != nil {
		return nil, err
	}
	return jsonResult(d)
}

func pushFile(ctx context.Context, in pushArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	out, err := android.PushFile(ctx, serial, in.LocalPath, in.DevicePath)
	if err != nil {
		return nil, err
	}
	return text("%s", strings.TrimSpace(out)), nil
}

func pullFile(ctx context.Context, in pullArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	out, err := android.PullFile(ctx, serial, in.DevicePath, in.LocalPath)
	if err != nil {
		return nil, err
	}
	return text("%s", strings.TrimSpace(out)), nil
}
