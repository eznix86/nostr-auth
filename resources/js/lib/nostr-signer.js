import { Nip07ExtensionSigner, Nip46RemoteSigner } from "nostr-signer-connector";

const NOSTR_CONNECT_SESSION_KEY = "nostr-connect-session";
const DEFAULT_NOSTR_CONNECT_RELAY = "wss://relay.nsec.app/";

export function getExtensionSigner() {
  if (typeof window === "undefined" || !window.nostr) {
    throw new Error("Nostr signer not found. Install a NIP-07 extension first.");
  }

  return new Nip07ExtensionSigner(window.nostr);
}

export async function resumeRemoteSignerSession() {
  if (typeof window === "undefined") {
    return null;
  }

  const rawSession = window.localStorage.getItem(NOSTR_CONNECT_SESSION_KEY);
  if (!rawSession) {
    return null;
  }

  try {
    return await Nip46RemoteSigner.resumeSession(JSON.parse(rawSession));
  } catch {
    clearRemoteSignerSession();
    return null;
  }
}

export function startRemoteSignerSession({ appName, authRelay }) {
  return Nip46RemoteSigner.listenConnectionFromRemote(buildRelayList(authRelay), {
    name: appName,
    url: typeof window === "undefined" ? "" : window.location.origin,
    description: `${appName} authentication`,
  });
}

export async function connectToBunker(bunkerUri) {
  return Nip46RemoteSigner.connectToRemote(bunkerUri);
}

export function persistRemoteSignerSession(session) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(NOSTR_CONNECT_SESSION_KEY, JSON.stringify(session));
}

export function clearRemoteSignerSession() {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.removeItem(NOSTR_CONNECT_SESSION_KEY);
}

function buildRelayList(authRelay) {
  const relayUrls = [DEFAULT_NOSTR_CONNECT_RELAY];

  if (typeof authRelay === "string" && authRelay.startsWith("ws")) {
    relayUrls.push(authRelay);
  }

  return [...new Set(relayUrls)];
}
