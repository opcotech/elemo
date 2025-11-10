import { Search } from "lucide-react";

import { Input } from "@/components/ui/input";

interface SearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
}

export function SearchInput({
  value,
  onChange,
  placeholder = "Search...",
  disabled = false,
  className,
}: SearchInputProps) {
  return (
    <div className={`relative max-w-md flex-1 ${className || ""}`}>
      <Search className="text-muted-foreground absolute top-3 left-2 h-4 w-4" />
      <Input
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        className="pl-8"
      />
    </div>
  );
}
