package platformrbac

import "testing"

func TestHasPermissionWildcard(t *testing.T) {
	if !HasPermission([]string{"operator"}, PermPlayersRead) {
		t.Fatal("operator should imply all permissions")
	}
}

func TestHasPermissionSupport(t *testing.T) {
	if !HasPermission([]string{"support"}, PermPlayersRead) {
		t.Fatal("support should read players")
	}
	if !HasPermission([]string{"support"}, PermEconomyRead) {
		t.Fatal("support should read economy ledger")
	}
	if HasPermission([]string{"support"}, PermSecurityWrite) {
		t.Fatal("support must not get security.write")
	}
	if !HasPermission([]string{"support"}, PermBackupsRestoreRequest) {
		t.Fatal("support should request restores")
	}
	if HasPermission([]string{"support"}, PermBackupsRestoreApprove) {
		t.Fatal("support must not approve restores by default")
	}
	if !HasPermission([]string{"security_admin"}, PermBackupsRestoreApprove) {
		t.Fatal("security_admin should approve restores")
	}
	if !HasPermission([]string{"security_admin"}, PermBackupsRestoreFulfill) {
		t.Fatal("security_admin should fulfill restores")
	}
}

func TestHasPermissionSecurityAdmin(t *testing.T) {
	if !HasPermission([]string{"security_admin"}, PermAuditRead) {
		t.Fatal("security_admin should read audit")
	}
	if !HasPermission([]string{"security_admin"}, PermSecurityWrite) {
		t.Fatal("security_admin should write security")
	}
	if !HasPermission([]string{"security_admin"}, PermEconomyRead) {
		t.Fatal("security_admin should read economy ledger")
	}
	if HasPermission([]string{"security_admin"}, PermSupportAck) {
		t.Fatal("security_admin should not get support ack by default")
	}
}
