import * as React from 'react';

import { Header } from '@/components/header';
import { Shell, ShellContent } from '@/components/ui/shell';

const Layout = ({ children }: Readonly<{ children: React.ReactNode }>) => {
  return (
    <Shell>
      <Header />
      <ShellContent>{children}</ShellContent>
    </Shell>
  );
};

export { Layout };
