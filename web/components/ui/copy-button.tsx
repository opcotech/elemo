import * as React from 'react';
import { IconClipboard } from '@tabler/icons-react';

import { cn } from '@/lib/utils';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

export interface CopyButtonProps extends Omit<React.HTMLAttributes<HTMLButtonElement>, 'onClick'> {
  targetRef: React.MutableRefObject<any>;
  className?: string;
  tooltipMessage?: string;
  tooltipDismissDelay?: number;
}

const CopyButton = React.forwardRef<HTMLButtonElement, CopyButtonProps>(
  ({ className, targetRef, tooltipMessage = 'copied', tooltipDismissDelay = 1500, ...props }, ref) => {
    const [copySuccess, setCopySuccess] = React.useState(false);

    function handleCopy(e: React.MouseEvent<HTMLElement>) {
      e.preventDefault();
      navigator.clipboard.writeText(targetRef.current.innerHTML);
      setCopySuccess(true);

      setTimeout(() => {
        setCopySuccess(false);
      }, tooltipDismissDelay);
    }

    return (
      <TooltipProvider>
        <Tooltip open={copySuccess}>
          <TooltipTrigger asChild>
            <button
              ref={ref}
              className={cn(
                'rounded-md p-1.5 text-foreground/70 hover:text-foreground focus:outline-none focus:ring-0 group-[.destructive]:text-red-300 group-[.destructive]:hover:text-red-50 group-[.destructive]:focus:ring-red-400 group-[.destructive]:focus:ring-offset-red-600',
                className
              )}
              onClick={handleCopy}
              {...props}
            >
              <IconClipboard className={cn('ml-auto h-4 w-4', copySuccess && 'text-accent')} />
            </button>
          </TooltipTrigger>
          <TooltipContent variant='blue'>{tooltipMessage}</TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }
);
CopyButton.displayName = 'CopyButton';

export { CopyButton };
