import { Deferred, Head, router } from "@inertiajs/react";
import { useState } from "react";

import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "../components/ui/card";

export default function Logout({ authenticatedPubkey, profile, title }) {
  const [error, setError] = useState("");
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const identityLabel = displayIdentity(profile, authenticatedPubkey);

  async function logout() {
    try {
      setError("");
      setIsLoggingOut(true);
      const csrfToken = await fetchCsrfToken();
      router.post("/logout", {}, { headers: { "X-CSRF-Token": csrfToken } });
    } catch (err) {
      setIsLoggingOut(false);
      setError(err instanceof Error ? err.message : "Failed to logout.");
    }
  }

  const helper = error || "You can continue now.";

  return (
    <>
      <Head title={title} />
      <Card className="w-full max-w-sm border-border bg-card/95 text-card-foreground shadow-[0_30px_80px_rgba(0,0,0,0.42)] backdrop-blur-md">
        <CardHeader className="gap-1.5">
          <CardTitle className="text-center text-xl">{title}</CardTitle>
          <CardDescription className="text-center">You are signed in with Nostr</CardDescription>
        </CardHeader>

        <CardContent className="flex flex-col gap-4">
          <Deferred data="profile" fallback={<ProfileSkeleton />}>
            {profile ? (
              <div className="flex min-w-0 flex-col items-center gap-2">
                {profile.picture ? <img src={profile.picture} alt={profile.name || profile.display_name} className="h-12 w-12 rounded-full" /> : null}
                <p className="max-w-full truncate text-sm font-medium" title={identityLabel}>{identityLabel}</p>
                {profile.nip05 ? <p className="max-w-full truncate text-xs text-muted-foreground" title={profile.nip05}>{profile.nip05}</p> : null}
              </div>
            ) : (
              <div className="min-w-0 text-center text-sm text-muted-foreground">
                Signed in as <code className="break-all" title={authenticatedPubkey}>{shortPubkey(authenticatedPubkey)}</code>
              </div>
            )}
          </Deferred>

          <p className={`text-center text-sm ${error ? "text-[#ffb4a9] dark:text-[#ffb4a9] light:text-[#a2362c]" : "text-muted-foreground"}`}>{helper}</p>
        </CardContent>

        <CardFooter>
          <Button className="w-full" variant="outline" onClick={logout} disabled={isLoggingOut}>
            {isLoggingOut ? "Logging out..." : "Logout"}
          </Button>
        </CardFooter>
      </Card>
    </>
  );
}

function displayIdentity(profile, authenticatedPubkey) {
  const displayName = profile?.display_name?.trim();
  if (displayName) {
    return shortenLabel(displayName);
  }

  const name = profile?.name?.trim();
  if (name) {
    return shortenLabel(name);
  }

  return shortPubkey(authenticatedPubkey);
}

function shortenLabel(value) {
  if (!value || value.length <= 32) {
    return value;
  }

  return `${value.slice(0, 14)}...${value.slice(-14)}`;
}

function shortPubkey(pubkey) {
  if (!pubkey || pubkey.length <= 28) {
    return pubkey;
  }

  return `${pubkey.slice(0, 14)}...${pubkey.slice(-14)}`;
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
function ProfileSkeleton() {
  return (
    <div className="flex flex-col items-center gap-2 py-2">
      <div className="h-12 w-12 animate-pulse rounded-full bg-muted" />
      <div className="h-4 w-24 animate-pulse rounded bg-muted" />
      <div className="h-3 w-32 animate-pulse rounded bg-muted" />
    </div>
  );
}
