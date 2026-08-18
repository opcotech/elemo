import type {
  SuggestionKeyDownProps,
  SuggestionProps,
} from "@tiptap/suggestion";

export interface SuggestionListItem {
  id: string;
  label: string;
  detail?: string;
  group?: string;
  icon?: string;
}

type SuggestionRendererProps<T> = SuggestionProps<T>;

function positionElement(
  element: HTMLElement,
  clientRect: (() => DOMRect | null) | null | undefined
) {
  const rect = clientRect?.();
  if (!rect) {
    return;
  }

  element.style.position = "fixed";
  element.style.left = `${Math.max(8, rect.left)}px`;
  element.style.top = `${rect.bottom + 6}px`;
  element.style.zIndex = "60";
}

export function createSuggestionListRenderer<T extends SuggestionListItem>(
  onSelect: (item: T) => Record<string, unknown> | T = (item) => item
) {
  return () => {
    let element: HTMLDivElement | null = null;
    let selectedIndex = 0;
    let currentProps: SuggestionRendererProps<T> | null = null;

    const renderItems = () => {
      if (!element || !currentProps) {
        return;
      }

      const items = currentProps.items;
      element.replaceChildren();
      element.className =
        "rich-text-suggestion border-border bg-popover text-popover-foreground shadow-md";

      if (items.length === 0) {
        const empty = document.createElement("div");
        empty.className = "rich-text-suggestion__empty";
        empty.textContent = "No matches";
        element.appendChild(empty);
        positionElement(element, currentProps.clientRect);
        return;
      }

      items.forEach((item, index) => {
        if (item.group && item.group !== items[index - 1]?.group) {
          const heading = document.createElement("div");
          heading.className = "rich-text-suggestion__group";
          heading.textContent = item.group;
          element?.appendChild(heading);
        }

        const button = document.createElement("button");
        button.type = "button";
        button.className = "rich-text-suggestion__item";
        if (index === selectedIndex) {
          button.classList.add("is-active");
        }

        if (item.icon) {
          const icon = document.createElement("span");
          icon.className = "rich-text-suggestion__icon";
          icon.textContent = item.icon;
          button.appendChild(icon);
        }

        const label = document.createElement("span");
        label.className = "rich-text-suggestion__label";
        label.textContent = item.label;
        button.appendChild(label);

        if (item.detail) {
          const detail = document.createElement("span");
          detail.className = "rich-text-suggestion__detail";
          detail.textContent = item.detail;
          button.appendChild(detail);
        }

        button.addEventListener("mousedown", (event) => {
          event.preventDefault();
          currentProps?.command(onSelect(item) as never);
        });

        element?.appendChild(button);
      });

      positionElement(element, currentProps.clientRect);
    };

    return {
      onStart(props: SuggestionRendererProps<T>) {
        element = document.createElement("div");
        document.body.appendChild(element);
        selectedIndex = 0;
        currentProps = props;
        renderItems();
      },

      onUpdate(props: SuggestionRendererProps<T>) {
        currentProps = props;
        selectedIndex = 0;
        renderItems();
      },

      onKeyDown({ event }: SuggestionKeyDownProps) {
        if (!currentProps) {
          return false;
        }

        const itemCount = currentProps.items.length;
        if (event.key === "ArrowUp") {
          if (itemCount === 0) {
            return true;
          }
          selectedIndex = (selectedIndex + itemCount - 1) % itemCount;
          renderItems();
          return true;
        }

        if (event.key === "ArrowDown") {
          if (itemCount === 0) {
            return true;
          }
          selectedIndex = (selectedIndex + 1) % itemCount;
          renderItems();
          return true;
        }

        if (event.key === "Enter") {
          const item = currentProps.items[selectedIndex];
          if (item) {
            currentProps.command(onSelect(item) as never);
          }
          return true;
        }

        if (event.key === "Escape") {
          element?.remove();
          element = null;
          return true;
        }

        return false;
      },

      onExit() {
        element?.remove();
        element = null;
        currentProps = null;
      },
    };
  };
}
