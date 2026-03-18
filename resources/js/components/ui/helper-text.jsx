import { cn } from "../../lib/utils";

export function HelperText({ error, children, className }) {
  return (
    <p
      className={cn(
        "text-center text-sm",
        error ? "text-[#ffb4a9] dark:text-[#ffb4a9] light:text-[#a2362c]" : "text-muted-foreground",
        className,
      )}
    >
      {children}
    </p>
  );
}
