package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aurumflow/config"
	"aurumflow/internal/execacct"
	"aurumflow/internal/logger"
)

func runResolveDemoAccount(ctx context.Context) {
	if !demoEnvConfigured() {
		logger.Error("resolve-demo-account: missing AURUMFLOW_DEMO_*")
		os.Exit(1)
	}
	sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled)
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
	ar, err := sess.client.GetAccounts(ctx)
	if err != nil {
		logger.Error("accounts: %v", err)
		os.Exit(1)
	}
	explicit := strings.TrimSpace(os.Getenv("AURUMFLOW_DEMO_ACCOUNT_ID"))
	ident := execacct.Resolve(ar.Accounts, explicit, "")
	fmt.Printf("resolution=%s accounts[0]_fallback=%s\n", ident.Resolution, execacct.Accounts0Fallback)
	fmt.Printf("verified=%v may_trade=%v reason=%s\n", ident.Verified, ident.MayTrade, ident.Reason)
	for i, a := range ar.Accounts {
		pref := ""
		if a.Preferred {
			pref = " preferred"
		}
		fmt.Printf("account[%d] id=%s type=%s currency=%s status=%s%s\n",
			i, execacct.Mask(a.AccountID), emptyDash(a.AccountType), emptyDash(a.Currency), emptyDash(a.Status), pref)
	}
	if ident.AccountID != "" {
		fmt.Printf("candidate id=%s type=%s currency=%s status=%s\n",
			execacct.Mask(ident.AccountID), emptyDash(ident.Type), emptyDash(ident.Currency), emptyDash(ident.Status))
		writeLocalAccountEnv(ident.AccountID)
	}
	if !ident.MayTrade {
		fmt.Println(execacct.Instruction(ident))
	}
}

// writeLocalAccountEnv stores the full account id only in a local temp file
// (0600). Stdout never contains the raw id — reports/git must stay masked.
func writeLocalAccountEnv(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	path := filepath.Join(os.TempDir(), "aurumflow-demo-account.id")
	if err := os.WriteFile(path, []byte(id+"\n"), 0o600); err != nil {
		fmt.Printf("local_env_file=WRITE_FAILED\n")
		return
	}
	fmt.Printf("local_env_file=%s last4=%s\n", path, execacct.Mask(id))
	fmt.Println("PowerShell: $env:AURUMFLOW_DEMO_ACCOUNT_ID = (Get-Content $env:TEMP\\aurumflow-demo-account.id).Trim()")
}
