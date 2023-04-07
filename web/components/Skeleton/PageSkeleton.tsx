import Page from '@/components/Page';

import ContentSkeleton from './ContentSkeleton';
import TextSkeleton from './TextSkeleton';

export default async function PageSkeleton() {
  return (
    <Page>
      <ContentSkeleton>
        <div className={'p-10'}>
          <TextSkeleton/>
        </div>
      </ContentSkeleton>
    </Page>
  );
}
