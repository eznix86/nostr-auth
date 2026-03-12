import * as React from "react";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import { X } from "lucide-react";

import { cn } from "../../lib/utils";

export function Dialog(props) {
  return <DialogPrimitive.Root data-slot="dialog" {...props} />;
}

export function DialogTrigger(props) {
  return <DialogPrimitive.Trigger data-slot="dialog-trigger" {...props} />;
}

export function DialogPortal(props) {
  return <DialogPrimitive.Portal data-slot="dialog-portal" {...props} />;
}

export function DialogClose(props) {
  return <DialogPrimitive.Close data-slot="dialog-close" {...props} />;
}

export function DialogOverlay({ className, ...props }) {
  return (
    <DialogPrimitive.Overlay
      data-slot="dialog-overlay"
      className={cn("fixed inset-0 z-50 bg-black/65 backdrop-blur-sm", className)}
      {...props}
    />
  );
}

export function DialogContent({ className, children, showCloseButton = true, ...props }) {
  return (
    <DialogPortal>
      <DialogOverlay />
      <DialogPrimitive.Content
        data-slot="dialog-content"
        className={cn(
          "fixed top-1/2 left-1/2 z-50 grid w-[calc(100%-2rem)] max-w-lg -translate-x-1/2 -translate-y-1/2 gap-4 rounded-2xl border border-border/80 bg-card/95 p-6 text-card-foreground shadow-[0_24px_90px_rgba(0,0,0,0.5)] backdrop-blur-xl duration-200 sm:w-full",
          className,
        )}
        {...props}
      >
        {children}

        {showCloseButton ? (
          <DialogPrimitive.Close
            className="absolute top-4 right-4 inline-flex h-9 w-9 items-center justify-center rounded-md border border-border/70 bg-background/70 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] outline-none"
          >
            <X className="h-4 w-4" />
            <span className="sr-only">Close</span>
          </DialogPrimitive.Close>
        ) : null}
      </DialogPrimitive.Content>
    </DialogPortal>
  );
}

export function DialogHeader({ className, ...props }) {
  return <div data-slot="dialog-header" className={cn("flex flex-col gap-1.5 text-left", className)} {...props} />;
}

export function DialogFooter({ className, ...props }) {
  return <div data-slot="dialog-footer" className={cn("flex flex-col-reverse gap-2 sm:flex-row sm:justify-end", className)} {...props} />;
}

export function DialogTitle({ className, ...props }) {
  return <DialogPrimitive.Title data-slot="dialog-title" className={cn("text-lg font-semibold", className)} {...props} />;
}

export function DialogDescription({ className, ...props }) {
  return <DialogPrimitive.Description data-slot="dialog-description" className={cn("text-sm text-muted-foreground", className)} {...props} />;
}
