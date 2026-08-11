import { useEffect, useState } from "react";

import { DocumentQuickCreate } from "@/components/quick-create/document";
import { TodoQuickCreate } from "@/components/quick-create/todo";
import {
  QUICK_CREATE_EVENT,
  isQuickCreateType,
  isTypingTarget,
} from "@/components/quick-create/types";
import type { QuickCreateType } from "@/components/quick-create/types";
import { WorkQuickCreate } from "@/components/quick-create/work";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface QuickCreateProps {
  initialOpen?: boolean;
  initialType?: QuickCreateType;
}

export function QuickCreate({
  initialOpen = false,
  initialType = "todo",
}: QuickCreateProps = {}) {
  const [open, setOpen] = useState(initialOpen);
  const [type, setType] = useState<QuickCreateType>(initialType);

  useEffect(() => {
    const openDialog = (event: Event) => {
      if (event instanceof CustomEvent && isQuickCreateType(event.detail)) {
        setType(event.detail);
      } else {
        setType("todo");
      }
      setOpen(true);
    };
    const handleKeyDown = (event: KeyboardEvent) => {
      if (
        event.key.toLowerCase() === "c" &&
        !event.metaKey &&
        !event.ctrlKey &&
        !event.altKey &&
        !isTypingTarget(event.target)
      ) {
        event.preventDefault();
        setType("todo");
        setOpen(true);
      }
    };
    window.addEventListener(QUICK_CREATE_EVENT, openDialog);
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener(QUICK_CREATE_EVENT, openDialog);
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  const handleOpenChange = (nextOpen: boolean) => {
    setOpen(nextOpen);
    if (!nextOpen) {
      setType("todo");
    }
  };

  const handleComplete = () => {
    handleOpenChange(false);
  };

  const handleCancel = () => {
    handleOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Quick create</DialogTitle>
          <DialogDescription>
            Create in the current context. Only the title is required.
          </DialogDescription>
        </DialogHeader>

        <div className="mt-5 flex w-full flex-col gap-2">
          <Label htmlFor="quick-create-type">Entity type</Label>
          <Select
            value={type}
            onValueChange={(value) => {
              if (isQuickCreateType(value)) {
                setType(value);
              }
            }}
            items={{
              todo: "Personal todo",
              work: "Work item",
              document: "Document",
            }}
          >
            <SelectTrigger id="quick-create-type" className="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="todo">Personal todo</SelectItem>
              <SelectItem value="work">Work item</SelectItem>
              <SelectItem value="document">Document</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {type === "todo" && (
          <TodoQuickCreate
            onCancel={handleCancel}
            onComplete={handleComplete}
          />
        )}
        {type === "work" && (
          <WorkQuickCreate
            onCancel={handleCancel}
            onComplete={handleComplete}
          />
        )}
        {type === "document" && (
          <DocumentQuickCreate
            onCancel={handleCancel}
            onComplete={handleComplete}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

export type { QuickCreateType } from "@/components/quick-create/types";
