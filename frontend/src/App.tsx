import { useEffect, useState } from "react";
import { AnimatedBackground, CopyButton, ToastProvider, useToast } from "@block52/ui-sdk";
import "@block52/ui-sdk/styles.css";
import {
    Address,
    GenerateWallet,
    RecentBlocks,
    SignQuery,
    WalletStatus,
} from "../wailsjs/go/main/App";
import "./App.css";

// The Block52 node this spike talks to. In the full app these come from a
// network picker; hardcoded here to prove the Go core end to end.
const RPC = "https://node1.block52.xyz/rpc/";

interface Block {
    height: number;
    time: string;
    hash: string;
    numTxs: number;
}

function Dashboard() {
    const toast = useToast();
    const [address, setAddress] = useState<string>("");
    const [mnemonic, setMnemonic] = useState<string>("");
    const [latest, setLatest] = useState<number>(0);
    const [blocks, setBlocks] = useState<Block[]>([]);
    const [sig, setSig] = useState<string>("");
    const [error, setError] = useState<string>("");

    useEffect(() => {
        WalletStatus()
            .then(status => {
                if (status.exists && !status.locked) return Address().then(setAddress);
            })
            .catch(() => undefined);
    }, []);

    useEffect(() => {
        let active = true;
        const tick = () =>
            RecentBlocks(RPC, 8)
                .then(res => {
                    if (!active) return;
                    setLatest(res.latest as number);
                    setBlocks((res.blocks as Block[]) ?? []);
                })
                .catch(e => active && setError(String(e)));
        tick();
        const t = setInterval(tick, 5000);
        return () => {
            active = false;
            clearInterval(t);
        };
    }, []);

    const onGenerate = async () => {
        try {
            const created = await GenerateWallet();
            setMnemonic(created.mnemonic);
            setAddress(created.address);
            setError("");
            toast.showSuccess("Wallet created in Go");
        } catch (e) {
            setError(String(e));
        }
    };

    const onSignQuery = async () => {
        try {
            setSig(await SignQuery(Math.floor(Date.now() / 1000)));
            toast.showSuccess("Signed in Go");
        } catch (e) {
            setError(String(e));
        }
    };

    return (
        <main className="content">
            <h1>Block52 — Wails + Go core</h1>
            <p className="sub">Wallet, signing and chain reads all run in Go (pokerchain-go-client).</p>

            <section className="pane">
                <h2>Wallet</h2>
                {address ? (
                    <div className="row">
                        <code className="addr">{address}</code>
                        <CopyButton text={address} />
                    </div>
                ) : (
                    <button onClick={onGenerate}>Generate wallet</button>
                )}
                {mnemonic && <p className="seed">Seed (write it down): {mnemonic}</p>}
                {address && (
                    <div className="row">
                        <button onClick={onSignQuery}>Sign table-query</button>
                        {sig && <code className="sig">{sig.slice(0, 22)}…</code>}
                    </div>
                )}
            </section>

            <section className="pane">
                <h2>Recent blocks {latest > 0 && <span className="muted">· tip {latest}</span>}</h2>
                <ul className="blocks">
                    {blocks.map(b => (
                        <li key={b.height}>
                            <span className="height">#{b.height}</span>
                            <span className="muted">{b.numTxs} tx</span>
                            <span className="muted small">{b.time.replace("T", " ").slice(0, 19)}</span>
                        </li>
                    ))}
                </ul>
            </section>

            {error && <p className="error">{error}</p>}
        </main>
    );
}

function App() {
    return (
        <div id="app">
            <AnimatedBackground />
            <ToastProvider>
                <Dashboard />
            </ToastProvider>
        </div>
    );
}

export default App;
