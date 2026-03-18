import { router } from "@inertiajs/react";

export async function fetchCsrfToken() {
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

export async function fetchChallenge() {
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

export function postAuthVerify(event, csrfToken, options = {}) {
  router.post(
    "/auth/verify",
    { event: JSON.stringify(event) },
    {
      preserveScroll: true,
      preserveState: false,
      headers: {
        "X-CSRF-Token": csrfToken,
      },
      ...options,
    },
  );
}

export function postLogout(csrfToken, options = {}) {
  router.post(
    "/logout",
    {},
    {
      headers: {
        "X-CSRF-Token": csrfToken,
      },
      ...options,
    },
  );
}
