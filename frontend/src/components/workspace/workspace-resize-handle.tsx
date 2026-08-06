import { GripVertical } from "lucide-react";
import { Separator } from "react-resizable-panels";

import { cn } from "@/lib/utils";

interface WorkspaceResizeHandleProps {
  label: string;
  className?: string;
}

export function WorkspaceResizeHandle({
  label,
  className,
}: WorkspaceResizeHandleProps) {
  return (
    <Separator
      aria-label={label}
      className={cn("workspace-resize-handle", className)}
    >
      <span aria-hidden="true" className="workspace-resize-handle__grip">
        <GripVertical size={11} strokeWidth={2} />
      </span>
    </Separator>
  );
}
