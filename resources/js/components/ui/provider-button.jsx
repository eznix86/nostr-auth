import { Loader2 } from "lucide-react";

import { Button } from "./button";
import { cn } from "../../lib/utils";

export function ProviderButton({ title, icon, onClick, loading, className, ...rest }) {
  return (
    <Button onClick={onClick} className={cn("w-full rounded-md", className)} variant="outline" {...rest}>
      {loading ? (
        <Loader2 className="animate-spin" />
      ) : (
        <>
          {icon}
          {title}
        </>
      )}
    </Button>
  );
}
