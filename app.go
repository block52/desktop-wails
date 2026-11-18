package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/block52/pokerchain-go-client/chain"
	"github.com/block52/pokerchain-go-client/keys"
	"github.com/block52/pokerchain-go-client/lightnode"
	"github.com/block52/pokerchain-go-client/wallet"
)

const addressPrefix = "b52"

// App is the Go core, bound to the React frontend by Wails. Everything the UI
// needs — wallet, signing, chain reads — is a method here; there is no separate
// sidecar process and no preload IPC. The wallet key lives only in this process.
type App struct {
	ctx     context.Context
	vault   *wallet.Vault
	windows map[string]*lightnode.Window
	dataDir string
}

func NewApp() *App {
	return &App{windows: map[string]*lightnode.Window{}}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.dataDir = appDataDir()
	a.vault = wallet.NewVault(filepath.Join(a.dataDir, "wallet-vault.json"), noopKeychain{}, 0)
}

// --- wallet ---------------------------------------------------------------

// WalletStatus mirrors the vault status for the UI.
type WalletStatus struct {
	Exists      bool `json:"exists"`
	Locked      bool `json:"locked"`
	HasPassword bool `json:"hasPassword"`
}

func (a *App) WalletStatus() (WalletStatus, error) {
	st, err := a.vault.Status()
	if err != nil {
		return WalletStatus{}, err
	}
	return WalletStatus{Exists: st.Exists, Locked: st.Locked, HasPassword: st.HasPassword}, nil
}

// GenerateWallet makes a new 24-word wallet, saves it sealed, and returns the
// mnemonic once (to be written down) plus the address.
func (a *App) GenerateWallet() (map[string]string, error) {
	mnemonic, err := wallet.NewMnemonic(256)
	if err != nil {
		return nil, err
	}
	if err := a.vault.SaveMnemonic(mnemonic); err != nil {
		return nil, err
	}
	addr, err := wallet.Address(mnemonic, addressPrefix)
	if err != nil {
		return nil, err
	}
	return map[string]string{"mnemonic": mnemonic, "address": addr}, nil
}

// ImportWallet saves a seed phrase the user typed and returns its address.
func (a *App) ImportWallet(mnemonic string) (string, error) {
	if !wallet.IsValidMnemonic(mnemonic) {
		return "", fmt.Errorf("invalid seed phrase")
	}
	if err := a.vault.SaveMnemonic(mnemonic); err != nil {
		return "", err
	}
	return wallet.Address(mnemonic, addressPrefix)
}

// Address returns the saved wallet's address (requires the vault be unlocked).
func (a *App) Address() (string, error) {
	mnemonic, err := a.vault.Mnemonic()
	if err != nil {
		return "", err
	}
	return wallet.Address(mnemonic, addressPrefix)
}

func (a *App) UnlockWithPassword(password string) (bool, error) {
	return a.vault.UnlockWithPassword(password)
}

func (a *App) SetPassword(password string) error { return a.vault.SetPassword(password) }

func (a *App) LockWallet() { a.vault.Lock() }

// SignQuery signs the table-subscription message pokerchain-query:<timestamp>
// (EIP-191), the same signature the table WebSocket checks.
func (a *App) SignQuery(timestamp int64) (string, error) {
	if timestamp <= 0 {
		return "", fmt.Errorf("invalid timestamp")
	}
	mnemonic, err := a.vault.Mnemonic()
	if err != nil {
		return "", err
	}
	priv, err := wallet.PrivateKey(mnemonic)
	if err != nil {
		return "", err
	}
	defer priv.Zero()
	return "0x" + hexEncode(keys.PersonalSign(priv, fmt.Sprintf("pokerchain-query:%d", timestamp))), nil
}

// --- lobby / chain reads --------------------------------------------------

// ListGames returns the node's games as the JSON string the chain emits.
func (a *App) ListGames(grpcEndpoint, chainID string) (string, error) {
	target, useTLS := parseGRPC(grpcEndpoint)
	c := chain.New(target, chainID, useTLS)
	ctx, cancel := withTimeout(a.ctx)
	defer cancel()
	return c.ListGames(ctx)
}

// Balance returns the micro-denom balance of an address, as a string.
func (a *App) Balance(grpcEndpoint, chainID, address, denom string) (string, error) {
	target, useTLS := parseGRPC(grpcEndpoint)
	c := chain.New(target, chainID, useTLS)
	ctx, cancel := withTimeout(a.ctx)
	defer cancel()
	return c.Balance(ctx, address, denom)
}

// RecentBlocks returns the newest `count` blocks from a rolling 1000-block
// window polled from the node's Tendermint RPC.
func (a *App) RecentBlocks(rpcEndpoint string, count int) (map[string]any, error) {
	win, ok := a.windows[rpcEndpoint]
	if !ok {
		win = lightnode.NewWindow(rpcEndpoint, 1000, nil)
		a.windows[rpcEndpoint] = win
	}
	ctx, cancel := withTimeout(a.ctx)
	defer cancel()
	if err := win.Refresh(ctx); err != nil && len(win.Blocks(0)) == 0 {
		return nil, err
	}
	return map[string]any{"latest": win.Latest(), "blocks": win.Blocks(count)}, nil
}
