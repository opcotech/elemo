'use client';

import type { ErrorInfo, ReactNode } from 'react';
import { Component } from 'react';

import Button from '@/components/Button';

export interface ErrorBoundaryProps {
  children: ReactNode;
}

export interface ErrorBoundaryState {
  hasError: boolean;
}

export default class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error({ error, errorInfo });
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className={'h-screen w-screen flex items-center'}>
          <div className={'max-w-xl mx-auto text-center'}>
            <h2 className={'mb-4'}>Something went wrong!</h2>
            <p className={'mb-10'}>
              The application has encountered an error and cannot continue. Please try again later.
            </p>
            <Button onClick={() => this.setState({ hasError: false })}>Try again</Button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
