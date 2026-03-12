import * as React from "react";
import * as TabsPrimitive from "@radix-ui/react-tabs";

import { cn } from "../../lib/utils";

export function Tabs({ className, ...props }) {
  return <TabsPrimitive.Root data-slot="tabs" className={cn("flex flex-col gap-4", className)} {...props} />;
}

export function TabsList({ className, ...props }) {
  return (
    <TabsPrimitive.List
      data-slot="tabs-list"
      className={cn("inline-flex h-10 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground", className)}
      {...props}
    />
  );
}

export function TabsTrigger({ className, ...props }) {
  return (
    <TabsPrimitive.Trigger
      data-slot="tabs-trigger"
      className={cn(
        "inline-flex min-w-0 flex-1 items-center justify-center rounded-md px-3 py-1.5 text-sm font-medium whitespace-nowrap transition-[color,box-shadow] disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] outline-none",
        className,
      )}
      {...props}
    />
  );
}

export function TabsContent({ className, ...props }) {
  return <TabsPrimitive.Content data-slot="tabs-content" className={cn("outline-none", className)} {...props} />;
}
