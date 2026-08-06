import { PanelLeftOpen, PanelRightOpen } from "lucide-react";

import { cn } from "@/lib/utils";

interface WorkspaceEdgeButtonProps {
  label: string;
  onClick: () => void;
  side: "left" | "right";
  className?: string;
}

export function WorkspaceEdgeButton({
  label,
  onClick,
  side,
  className,
}: WorkspaceEdgeButtonProps) {
  const Icon = side === "left" ? PanelLeftOpen : PanelRightOpen;

  return (
    <button
      aria-label={label}
      className={cn(
        "absolute top-1/2 z-40 inline-flex h-10 w-7 -translate-y-1/2 items-center justify-center border border-[#d5d1ca] bg-[rgba(251,250,248,0.96)] text-[#5f6e68] shadow-[0_4px_16px_rgba(35,33,31,0.1)] backdrop-blur-md transition-[background-color,color] hover:bg-white hover:text-[#315c4d] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#527b6d]",
        side === "left"
          ? "left-0 rounded-r-md border-l-0"
          : "right-0 rounded-l-md border-r-0",
        className,
      )}
      onClick={onClick}
      title={label}
      type="button"
    >
      <Icon size={15} strokeWidth={1.8} />
    </button>
  );
}
