export function ProfileSummary({ profile, authenticatedPubkey, identityLabel }) {
  if (!profile) {
    return (
      <div className="min-w-0 text-center text-sm text-muted-foreground">
        Signed in as <code className="break-all" title={authenticatedPubkey}>{shortPubkey(authenticatedPubkey)}</code>
      </div>
    );
  }

  return (
    <div className="flex min-w-0 flex-col items-center gap-2">
      {profile.picture ? <img src={profile.picture} alt={profile.name || profile.display_name} className="h-12 w-12 rounded-full" /> : null}
      <p className="max-w-full truncate text-sm font-medium" title={identityLabel}>{identityLabel}</p>
      {profile.nip05 ? <p className="max-w-full truncate text-xs text-muted-foreground" title={profile.nip05}>{profile.nip05}</p> : null}
    </div>
  );
}

function shortPubkey(pubkey) {
  if (!pubkey || pubkey.length <= 28) {
    return pubkey;
  }

  return `${pubkey.slice(0, 14)}...${pubkey.slice(-14)}`;
}
