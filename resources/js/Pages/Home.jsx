import { Head, router, Deferred } from "@inertiajs/react";
import { useEffect, useMemo, useState } from "react";
import { Moon, Sun } from "lucide-react";

import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "../components/ui/card";
import { ProviderButton } from "../components/ui/provider-button";

export default function Home({ authenticatedPubkey, authError, challenge, csrfToken, redirectTo, relay, title, profile }) {
  const [status, setStatus] = useState(authenticatedPubkey ? "authenticated" : "idle");
  const [error, setError] = useState("");
  const [theme, setTheme] = useState(() => getInitialTheme());

  useEffect(() => {
    applyTheme(theme);
    window.localStorage.setItem("nostr-auth-theme", theme);
  }, [theme]);

  useEffect(() => {
    setStatus(authenticatedPubkey ? "authenticated" : "idle");
  }, [authenticatedPubkey]);

  const authEvent = useMemo(() => {
    return {
      kind: 22242,
      created_at: Math.floor(Date.now() / 1000),
      content: "",
      tags: [
        ["relay", relay],
        ["challenge", challenge],
      ],
    };
  }, [challenge, relay]);

  async function signChallenge() {
    if (!window.nostr) {
      setError("Nostr signer not found. Install a NIP-07 extension first.");
      setStatus("error");
      return;
    }

    try {
      setStatus("signing");
      setError("");

      const pubkey = await window.nostr.getPublicKey();
      const event = await window.nostr.signEvent({
        ...authEvent,
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

  function logout() {
    router.post("/logout", {}, { headers: { "X-CSRF-Token": csrfToken } });
  }

  const subtitle = authenticatedPubkey ? "You are signed in with Nostr" : "Welcome back, login with";
  const helper = error || authError ? error || authError.replaceAll("_", " ") : redirectTo ? "You will return to your app after signing in." : statusLabel(status);

  return (
    <>
      <Head title={title} />
      <div className="relative flex min-h-svh flex-col items-center justify-center overflow-hidden px-4">
        <div className="auth-bg-image absolute inset-0 bg-cover bg-center bg-no-repeat" />

        <div className="absolute inset-0 bg-black/8" />

        <div className="absolute top-4 right-4 z-20 flex flex-row gap-2">
          <Button className="relative border-border bg-card text-card-foreground shadow-sm hover:bg-card/90 dark:bg-card dark:hover:bg-card/90" size="icon" variant="outline" onClick={() => setTheme((current) => (current === "light" ? "dark" : "light"))}>
            <Sun className="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 transition-all dark:scale-0 dark:-rotate-90" />
            <Moon className="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 transition-all dark:scale-100 dark:rotate-0" />
            <span className="sr-only">Toggle theme</span>
          </Button>
        </div>

        <main className="relative z-10 flex min-h-svh w-full items-center justify-center py-20">
          <Card className="w-full max-w-sm border-border bg-card/95 text-card-foreground shadow-[0_30px_80px_rgba(0,0,0,0.42)] backdrop-blur-md">
            <CardHeader className="gap-1.5">
              <CardTitle className="text-center text-xl">{title}</CardTitle>
              <CardDescription className="text-center">{subtitle}</CardDescription>
            </CardHeader>

            <CardContent className="flex flex-col gap-4">
              {authenticatedPubkey ? (
                <Deferred data="profile" fallback={<ProfileSkeleton />}>
                  {profile && (
                    <div className="flex flex-col items-center gap-2">
                      {profile.picture && (
                        <img src={profile.picture} alt={profile.name || profile.display_name} className="h-12 w-12 rounded-full" />
                      )}
                      <p className="text-sm font-medium">{profile.display_name || profile.name}</p>
                      {profile.nip05 && <p className="text-xs text-muted-foreground">{profile.nip05}</p>}
                    </div>
                  )}
                </Deferred>
              ) : null}

              {!authenticatedPubkey ? (
                <div className="flex flex-col items-center justify-center gap-2.5">
                  <ProviderButton
                    title="Sign in with Nostr"
                    icon={<NostrMark />}
                    className="w-full"
                    onClick={signChallenge}
                    loading={status === "signing" || status === "verifying"}
                    disabled={status === "signing" || status === "verifying"}
                  />
                </div>
              ) : null}

              <p className={`text-center text-sm ${error || authError ? "text-[#ffb4a9] dark:text-[#ffb4a9] light:text-[#a2362c]" : "text-muted-foreground"}`}>{helper}</p>
            </CardContent>

            {authenticatedPubkey ? (
              <CardFooter>
                <Button className="w-full" variant="outline" onClick={logout}>
                  Logout
                </Button>
              </CardFooter>
            ) : null}
          </Card>
        </main>
      </div>
    </>
  );
}

function statusLabel(status) {
  switch (status) {
    case "authenticated":
      return "You can continue now.";
    case "signing":
      return "Waiting for your signer...";
    case "verifying":
      return "Verifying your signature...";
    case "error":
      return "Something went wrong.";
    default:
      return "Ready when you are.";
  }
}

function getInitialTheme() {
  if (typeof window === "undefined") {
    return "dark";
  }

  const storedTheme = window.localStorage.getItem("nostr-auth-theme");
  if (storedTheme === "light" || storedTheme === "dark") {
    return storedTheme;
  }

  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function applyTheme(theme) {
  document.documentElement.classList.toggle("dark", theme === "dark");
  document.documentElement.classList.toggle("light", theme === "light");
}

function NostrMark() {
  return (
    <svg viewBox="0 0 256 256" aria-hidden="true" className="h-[1.1rem] w-[1.1rem]" fill="currentColor">
      <path d="M210.8 199.4c0 3.1-2.5 5.7-5.7 5.7h-68c-3.1 0-5.7-2.5-5.7-5.7v-15.5c.3-19 2.3-37.2 6.5-45.5 2.5-5 6.7-7.7 11.5-9.1 9.1-2.7 24.9-.9 31.7-1.2 0 0 20.4.8 20.4-10.7s-9.1-8.6-9.1-8.6c-10 .3-17.7-.4-22.6-2.4-8.3-3.3-8.6-9.2-8.6-11.2-.4-23.1-34.5-25.9-64.5-20.1-32.8 6.2.4 53.3.4 116.1v8.4c0 3.1-2.6 5.6-5.7 5.6H57.7c-3.1 0-5.7-2.5-5.7-5.7v-144c0-3.1 2.5-5.7 5.7-5.7h31.7c3.1 0 5.7 2.5 5.7 5.7 0 4.7 5.2 7.2 9 4.5 11.4-8.2 26-12.5 42.4-12.5 36.6 0 64.4 21.4 64.4 68.7v83.2ZM150 99.3c0-6.7-5.4-12.1-12.1-12.1s-12.1 5.4-12.1 12.1 5.4 12.1 12.1 12.1S150 106 150 99.3Z" />
    </svg>
  );
}

function ProfileSkeleton() {
  return (
    <div className="flex flex-col items-center gap-2 py-2">
      <div className="h-12 w-12 rounded-full bg-muted animate-pulse" />
      <div className="h-4 w-24 rounded bg-muted animate-pulse" />
      <div className="h-3 w-32 rounded bg-muted animate-pulse" />
    </div>
  );
}
