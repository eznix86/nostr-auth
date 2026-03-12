import { Head, router } from "@inertiajs/react";
import { useState } from "react";
import { Check, Copy, Link, QrCode } from "lucide-react";
import { QRCodeSVG } from "qrcode.react";

import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "../components/ui/dialog";
import { ProviderButton } from "../components/ui/provider-button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../components/ui/tabs";
import { clearRemoteSignerSession, connectToBunker, getExtensionSigner, persistRemoteSignerSession, resumeRemoteSignerSession, startRemoteSignerSession } from "../lib/nostr-signer";

export default function Home({ authError, redirectTo, title }) {
  const [status, setStatus] = useState("idle");
  const [error, setError] = useState("");
  const [connectUri, setConnectUri] = useState("");
  const [bunkerUri, setBunkerUri] = useState("");
  const [copyState, setCopyState] = useState("idle");
  const [isConnectDialogOpen, setIsConnectDialogOpen] = useState(false);
  const [connectMode, setConnectMode] = useState("nostrconnect");

  async function signChallengeWithSigner(signer) {
    try {
      setStatus("signing");
      setError("");

      const [csrfToken, challenge] = await Promise.all([fetchCsrfToken(), fetchChallenge()]);
      const pubkey = await signer.getPublicKey();
      const event = await signer.signEvent({
        kind: 22242,
        created_at: Math.floor(Date.now() / 1000),
        content: "",
        tags: [
          ["relay", challenge.relay],
          ["challenge", challenge.token],
        ],
        pubkey,
      });

      setStatus("verifying");

      router.post(
        "/auth/verify",
        { event: JSON.stringify(event), redirectTo },
        {
          preserveScroll: true,
          preserveState: false,
          headers: {
            "X-CSRF-Token": csrfToken,
          },
          onError: () => {
            setStatus("error");
            setError("Verification failed.");
          },
        },
      );
    } catch (err) {
      setStatus("error");
      setError(err instanceof Error ? err.message : "Failed to sign challenge.");
    }
  }

  async function signWithExtension() {
    try {
      await signChallengeWithSigner(getExtensionSigner());
    } catch (err) {
      setStatus("error");
      setError(err instanceof Error ? err.message : "Failed to load your Nostr signer.");
    }
  }

  async function signWithNostrConnect() {
    setIsConnectDialogOpen(true);
    setConnectMode("nostrconnect");
    setError("");

    try {
      const restoredSigner = await resumeRemoteSignerSession();
      if (restoredSigner) {
        await signChallengeWithSigner(restoredSigner);
      }
    } catch (err) {
      setStatus("error");
      setError(err instanceof Error ? err.message : "Failed to load your Nostr signer.");
    }
  }

  async function generateNostrConnect() {
    try {
      setError("");
      setCopyState("idle");
      setConnectMode("nostrconnect");
      setConnectUri("");

      const challenge = await fetchChallenge();
      const { connectUri: nextConnectUri, established } = startRemoteSignerSession({
        appName: title,
        authRelay: challenge.relay,
      });

      setConnectUri(nextConnectUri);
      setStatus("connecting");

      const { signer, session } = await established;

      persistRemoteSignerSession(session);
      await signChallengeWithSigner(signer);
    } catch (err) {
      clearRemoteSignerSession();
      setStatus("error");
      setError(err instanceof Error ? err.message : "Failed to connect to your Nostr signer.");
    }
  }

  async function signWithBunker() {
    try {
      const trimmedBunkerUri = bunkerUri.trim();

      if (!trimmedBunkerUri) {
        setStatus("error");
        setError("Paste a bunker:// URI from your remote signer.");
        return;
      }

      if (!trimmedBunkerUri.startsWith("bunker://")) {
        setStatus("error");
        setError("The bunker connection must start with bunker://.");
        return;
      }

      setError("");
      setCopyState("idle");
      setStatus("connecting");

      const { signer, session } = await connectToBunker(trimmedBunkerUri);

      persistRemoteSignerSession(session);
      await signChallengeWithSigner(signer);
    } catch (err) {
      clearRemoteSignerSession();
      setStatus("error");
      setError(err instanceof Error ? err.message : "Failed to connect to your bunker signer.");
    }
  }

  async function copyConnectUri() {
    if (!connectUri || typeof navigator === "undefined" || !navigator.clipboard) {
      return;
    }

    try {
      await navigator.clipboard.writeText(connectUri);
      setCopyState("copied");

      window.setTimeout(() => {
        setCopyState("idle");
      }, 2000);
    } catch {
      setError("Could not copy the NostrConnect URI.");
    }
  }

  function resetNostrConnect() {
    clearRemoteSignerSession();
    setConnectUri("");
    setBunkerUri("");
    setCopyState("idle");
    setError("");
    setStatus("idle");
  }

  async function restartNostrConnect() {
    resetNostrConnect();
    await generateNostrConnect();
  }

  function handleConnectDialogChange(open) {
    setIsConnectDialogOpen(open);

    if (!open) {
      resetNostrConnect();
    }
  }

  const helper = error || authError ? error || authError.replaceAll("_", " ") : redirectTo ? "You will return to your app after signing in." : statusLabel(status);

  return (
    <>
      <Head title={title} />
      <Card className="w-full max-w-sm border-border bg-card/95 text-card-foreground shadow-[0_30px_80px_rgba(0,0,0,0.42)] backdrop-blur-md">
        <CardHeader className="gap-1.5">
          <CardTitle className="text-center text-xl">{title}</CardTitle>
          <CardDescription className="text-center">Welcome back, login with</CardDescription>
        </CardHeader>

        <CardContent className="flex flex-col gap-4">
          <div className="flex flex-col items-center justify-center gap-2.5">
            <ProviderButton
              title="Sign in with Nostr"
              icon={<NostrMark />}
              className="w-full"
              onClick={signWithExtension}
              loading={status === "signing" || status === "verifying"}
              disabled={status === "connecting" || status === "signing" || status === "verifying"}
            />
            <ProviderButton
              title="Sign in with NostrConnect"
              icon={<QrCode />}
              className="w-full"
              onClick={signWithNostrConnect}
              loading={status === "connecting"}
              disabled={status === "connecting" || status === "signing" || status === "verifying"}
            />
          </div>

          <p className={`text-center text-sm ${error || authError ? "text-[#ffb4a9] dark:text-[#ffb4a9] light:text-[#a2362c]" : "text-muted-foreground"}`}>{helper}</p>
        </CardContent>
      </Card>

      <Dialog open={isConnectDialogOpen} onOpenChange={handleConnectDialogChange}>
        <DialogContent className="max-w-xl">
          <DialogHeader className="pr-10">
            <DialogTitle>Sign in with NostrConnect</DialogTitle>
            <DialogDescription>Choose a generated `nostrconnect://` code or paste a `bunker://` URI from your signer.</DialogDescription>
          </DialogHeader>

          <Tabs value={connectMode} onValueChange={setConnectMode} className="gap-5">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="nostrconnect">NostrConnect</TabsTrigger>
              <TabsTrigger value="bunker">bunker://</TabsTrigger>
            </TabsList>

            <TabsContent value="nostrconnect">
              <div className="grid gap-5 sm:grid-cols-[180px_minmax(0,1fr)] sm:items-start">
                <div className="mx-auto flex h-44 w-44 items-center justify-center rounded-2xl border border-border/80 bg-white p-3 shadow-sm">
                  {connectUri ? <QRCodeSVG bgColor="#ffffff" fgColor="#111827" includeMargin level="M" size={152} value={connectUri} /> : <QrCode className="h-10 w-10 text-slate-300" />}
                </div>

                <div className="min-w-0 space-y-3">
                  <div className="rounded-xl border border-border/70 bg-background/70 p-4">
                    <p className="mb-2 text-[11px] font-medium uppercase tracking-[0.18em] text-muted-foreground">NostrConnect URI</p>
                    <code className="block max-h-64 overflow-auto break-all text-xs leading-6 text-foreground">{connectUri || "Preparing secure connection..."}</code>
                  </div>

                  <Button className="w-full" variant="outline" onClick={copyConnectUri} type="button" disabled={!connectUri}>
                    {copyState === "copied" ? <Check /> : <Copy />}
                    {copyState === "copied" ? "Copied" : "Copy URI"}
                  </Button>

                  <Button className="w-full" onClick={generateNostrConnect} type="button" disabled={status === "connecting" || status === "signing" || status === "verifying"}>
                    <QrCode />
                    {connectUri ? "Generate new code" : "Generate code"}
                  </Button>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="bunker">
              <div className="space-y-3">
                <div className="rounded-xl border border-border/70 bg-background/70 p-4">
                  <label htmlFor="bunker-uri" className="mb-2 block text-[11px] font-medium uppercase tracking-[0.18em] text-muted-foreground">
                    bunker:// URI
                  </label>
                  <textarea
                    id="bunker-uri"
                    className="min-h-36 w-full resize-y rounded-lg border border-border/70 bg-background px-3 py-2 text-sm text-foreground outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]"
                    placeholder="bunker://pubkey?relay=wss%3A%2F%2F...&secret=..."
                    value={bunkerUri}
                    onChange={(event) => setBunkerUri(event.target.value)}
                  />
                </div>

                <Button className="w-full" onClick={signWithBunker} type="button" disabled={status === "connecting" || status === "signing" || status === "verifying"}>
                  <Link />
                  Connect bunker signer
                </Button>
              </div>
            </TabsContent>
          </Tabs>

          <DialogFooter className="items-stretch sm:items-center sm:justify-between">
            <p className="text-sm text-muted-foreground">{status === "connecting" ? (connectMode === "bunker" ? "Connecting to your bunker signer..." : "Waiting for your NostrConnect signer...") : helper}</p>
            {connectMode === "nostrconnect" ? (
              <Button variant="outline" onClick={restartNostrConnect} type="button" disabled={status === "connecting" || status === "signing" || status === "verifying"}>
                Generate new code
              </Button>
            ) : (
              <Button variant="outline" onClick={resetNostrConnect} type="button" disabled={status === "connecting" || status === "signing" || status === "verifying"}>
                Clear bunker URI
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

async function fetchCsrfToken() {
  const response = await fetch("/auth/csrf", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    throw new Error("Failed to refresh CSRF token.");
  }

  const payload = await response.json();
  if (!payload?.token) {
    throw new Error("Missing CSRF token.");
  }

  return payload.token;
}

async function fetchChallenge() {
  const response = await fetch("/auth/challenge", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    throw new Error("Failed to refresh login challenge.");
  }

  const payload = await response.json();
  if (!payload?.token || !payload?.relay) {
    throw new Error("Missing login challenge.");
  }

  return payload;
}

function statusLabel(status) {
  switch (status) {
    case "signing":
      return "Waiting for your signer...";
    case "verifying":
      return "Verifying your signature...";
    case "connecting":
      return "Waiting for your NostrConnect signer...";
    case "error":
      return "Something went wrong.";
    default:
      return "Ready when you are.";
  }
}
function NostrMark() {
  return (
    <svg viewBox="0 0 256 256" aria-hidden="true" className="h-[1.1rem] w-[1.1rem]" fill="currentColor">
      <path d="M210.8 199.4c0 3.1-2.5 5.7-5.7 5.7h-68c-3.1 0-5.7-2.5-5.7-5.7v-15.5c.3-19 2.3-37.2 6.5-45.5 2.5-5 6.7-7.7 11.5-9.1 9.1-2.7 24.9-.9 31.7-1.2 0 0 20.4.8 20.4-10.7s-9.1-8.6-9.1-8.6c-10 .3-17.7-.4-22.6-2.4-8.3-3.3-8.6-9.2-8.6-11.2-.4-23.1-34.5-25.9-64.5-20.1-32.8 6.2.4 53.3.4 116.1v8.4c0 3.1-2.6 5.6-5.7 5.6H57.7c-3.1 0-5.7-2.5-5.7-5.7v-144c0-3.1 2.5-5.7 5.7-5.7h31.7c3.1 0 5.7 2.5 5.7 5.7 0 4.7 5.2 7.2 9 4.5 11.4-8.2 26-12.5 42.4-12.5 36.6 0 64.4 21.4 64.4 68.7v83.2ZM150 99.3c0-6.7-5.4-12.1-12.1-12.1s-12.1 5.4-12.1 12.1 5.4 12.1 12.1 12.1S150 106 150 99.3Z" />
    </svg>
  );
}
