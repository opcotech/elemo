'use client';

import Message from '@/components/Message/index';
import useStore from '@/store';

export default function MessageArea() {
  const messages = useStore((state) => state.messages);

  return (
    <div
      aria-live="assertive"
      className="z-50 pointer-events-none fixed inset-0 flex items-end px-4 py-6 sm:items-start sm:pt-5 sm:pb-6"
    >
      <div className="flex w-full flex-col items-center space-y-4 sm:items-end">
        {messages.map((message, index) => {
          {
            /* TODO: We are not removing the message from the messages list */
          }
          return <Message key={index} {...message} />;
        })}
      </div>
    </div>
  );
}
