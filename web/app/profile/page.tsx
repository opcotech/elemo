import Page from "@/components/Page"
import {ContentSkeleton} from "@/components/Skeleton";

export const metadata = {
  title: 'Profile | Elemo',
}

export default function ProfilePage() {
  return (
    <Page title="Profile">
      <ContentSkeleton/>
    </Page>
  )
}
