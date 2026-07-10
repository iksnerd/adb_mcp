package android

import "context"

// GrantPermission grants a runtime permission, skipping its in-app dialog.
func GrantPermission(ctx context.Context, serial, pkg, permission string) error {
	_, err := runAdb(ctx, serial, "shell", "pm", "grant", pkg, permission)
	return err
}

// RevokePermission revokes a previously granted runtime permission.
func RevokePermission(ctx context.Context, serial, pkg, permission string) error {
	_, err := runAdb(ctx, serial, "shell", "pm", "revoke", pkg, permission)
	return err
}
