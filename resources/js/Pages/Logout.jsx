import { Deferred, Head, usePage } from "@inertiajs/react";
import { useState } from "react";

import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "../components/ui/card";
import { HelperText } from "../components/ui/helper-text";
import { ProfileSummary } from "../components/ui/profile-summary";
import { fetchCsrfToken, postLogout } from "../lib/auth-client";

export default function Logout() {
  const { props } = usePage();
  const { authenticatedPubkey, profile, title } = props;
  const [error, setError] = useState("");
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const identityLabel = displayIdentity(profile, authenticatedPubkey);

  async function logout() {
    try {
      setError("");
      setIsLoggingOut(true);
      const csrfToken = await fetchCsrfToken();
      postLogout(csrfToken, {
        onFinish: () => setIsLoggingOut(false),
        onError: () => setError("Failed to logout."),
      });
    } catch (err) {
      setIsLoggingOut(false);
      setError(err instanceof Error ? err.message : "Failed to logout.");
    }
  }

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
            <ProfileSummary profile={profile} authenticatedPubkey={authenticatedPubkey} identityLabel={identityLabel} />
          </Deferred>

          <HelperText error={Boolean(error)}>{error || "You can continue now."}</HelperText>
        </CardContent>

        <CardFooter>
          <Button className="w-full" variant="outline" onClick={logout} disabled={isLoggingOut} loading={isLoggingOut}>
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

function ProfileSkeleton() {
  return (
    <div className="flex flex-col items-center gap-2 py-2">
      <div className="h-12 w-12 animate-pulse rounded-full bg-muted" />
      <div className="h-4 w-24 animate-pulse rounded bg-muted" />
      <div className="h-3 w-32 animate-pulse rounded bg-muted" />
    </div>
  );
}
