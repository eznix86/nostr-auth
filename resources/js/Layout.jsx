import { usePage } from "@inertiajs/react";
import { useEffect, useState } from "react";
import { Moon, Sun } from "lucide-react";

import Background from "./components/Background";
import { Button } from "./components/ui/button";

export default function Layout({ children }) {
  const [theme, setTheme] = useState(() => getInitialTheme());
  const { branding } = usePage().props;

  useEffect(() => {
    applyTheme(theme);
    window.localStorage.setItem("nostr-auth-theme", theme);
  }, [theme]);

  return (
    <div className="relative flex min-h-svh flex-col items-center justify-center overflow-hidden px-4">
      <Background branding={branding} />
      <div className="absolute inset-0 bg-black/8" />

      <div className="absolute top-4 right-4 z-20 flex flex-row gap-2">
        <Button className="relative border-border bg-card text-card-foreground shadow-sm hover:bg-card/90 dark:bg-card dark:hover:bg-card/90" size="icon" variant="outline" onClick={() => setTheme((current) => (current === "light" ? "dark" : "light"))}>
          <Sun className="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 transition-all dark:scale-0 dark:-rotate-90" />
          <Moon className="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 transition-all dark:scale-100 dark:rotate-0" />
          <span className="sr-only">Toggle theme</span>
        </Button>
      </div>

      <main className="relative z-10 flex min-h-svh w-full items-center justify-center py-20">{children}</main>
    </div>
  );
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
