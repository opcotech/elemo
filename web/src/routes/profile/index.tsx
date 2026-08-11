import { createFileRoute } from '@tanstack/react-router'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'

export const Route = createFileRoute('/profile/')({
  component: ProfilePage,
})

function ProfilePage() {
  // mock data for now
  const user = {
    name: 'Demo User',
    username: 'demo',
    role: 'Senior Software Developer',
    email: 'demo@elemo.app',
    memberSince: 'August 7, 2025',
    status: 'active',
    languages: ['en', 'es'],
    phone: '+12345678900',
    location: '2900 S Congress Ave, Austin, TX',
    links: ['example.com', 'elemo.app'],
  }

  const orgMemberships = [
    {
      name: 'Acme Corporation',
      email: 'info@acme.com',
      website: 'acme.com',
      members: 3,
      joined: 'Jan 2024',
      role: 'Admin',
    },
    {
      name: 'TechStart Inc',
      email: 'hello@techstart.io',
      website: 'techstart.io',
      members: 12,
      joined: 'Mar 2024',
      role: 'Developer',
    },
  ]

  const recentActivity = [
    {
      title: 'Profile Updated',
      description: 'Updated bio and contact information',
      time: '2 hours ago',
      color: 'bg-blue-500',
    },
    {
      title: 'Joined Project',
      description: 'Added as team member with Developer role',
      time: '1 day ago',
      color: 'bg-blue-500',
    },
    {
      title: 'Task Completed',
      description: 'Successfully completed responsive design',
      time: '3 days ago',
      color: 'bg-green-500',
    },
  ]

  return (
    <div className="p-6 space-y-4 lg:flex lg:gap-6">
      {/* Left: main content */}
      <div className="flex-1 space-y-4">
        {/* Top profile card */}
        <Card className="p-6 space-y-4">
          {/* Breadcrumb + edit (breadcrumb is static for now) */}
          <div className="flex items-center justify-between text-sm text-muted-foreground mb-4">
            <div>
              <span className="text-muted-foreground">Profile</span>
              <span className="mx-2">/</span>
              <span className="font-medium text-foreground">{user.name}</span>
            </div>
            <Button variant="outline" size="sm">
              Edit Profile
            </Button>
          </div>

          <div className="flex flex-col gap-6 md:flex-row md:items-start">
            {/* Avatar placeholder */}
            <div className="h-20 w-20 rounded-full bg-muted" />

            <div className="flex-1 space-y-2">
              <div className="flex flex-wrap items-center gap-2">
                <h1 className="text-2xl font-semibold">{user.name}</h1>
                <Badge variant="outline" className="capitalize">
                  {user.status}
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground">@{user.username}</p>
              <p className="text-sm">{user.role}</p>
              <p className="text-sm text-muted-foreground">
                Member since {user.memberSince}
              </p>
            </div>
          </div>

          <Separator className="my-4" />

          {/* About + contact + extra details */}
          <div className="grid gap-6 md:grid-cols-3">
            <div className="space-y-2 md:col-span-1">
              <h2 className="text-sm font-semibold">About</h2>
              <p className="text-sm text-muted-foreground">
                Hello. It&apos;s me!
              </p>
            </div>

            <div className="space-y-2">
              <h2 className="text-sm font-semibold">Contact Information</h2>
              <div className="space-y-1 text-sm text-muted-foreground">
                <p>{user.email}</p>
                <p>{user.phone}</p>
                <p>{user.location}</p>
              </div>
            </div>

            <div className="space-y-2">
              <h2 className="text-sm font-semibold">Additional Details</h2>
              <div className="space-y-1 text-sm text-muted-foreground">
                <div>
                  <span className="font-medium text-foreground">Languages</span>
                  <div className="mt-1 flex flex-wrap gap-1">
                    {user.languages.map((lang) => (
                      <Badge key={lang} variant="outline" className="text-xs">
                        {lang}
                      </Badge>
                    ))}
                  </div>
                </div>
                <div className="mt-3">
                  <span className="font-medium text-foreground">Links</span>
                  <div className="mt-1 flex flex-wrap gap-1">
                    {user.links.map((link) => (
                      <Badge key={link} variant="outline" className="text-xs">
                        {link}
                      </Badge>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Card>

        {/* Organization memberships */}
        <Card className="p-6 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-sm font-semibold">
                Organization Memberships
              </h2>
              <p className="text-xs text-muted-foreground">
                Organizations and teams you&apos;re part of
              </p>
            </div>
          </div>

          <div className="space-y-4">
            {orgMemberships.map((org) => (
              <div key={org.name}>
                <div className="flex items-start justify-between gap-4">
                  <div className="space-y-1">
                    <p className="font-medium">{org.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {org.email}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {org.website}
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <Badge variant="outline" className="text-xs">
                      {org.role}
                    </Badge>
                    <Button variant="outline" size="sm">
                      View Organization
                    </Button>
                  </div>
                </div>
                <div className="mt-2 flex flex-wrap gap-4 text-xs text-muted-foreground">
                  <span>{org.members} members</span>
                  <span>Joined {org.joined}</span>
                </div>
                <Separator className="my-4 last:hidden" />
              </div>
            ))}
          </div>
        </Card>
      </div>

      {/* Right: recent activity */}
      <div className="mt-4 w-full space-y-4 lg:mt-0 lg:w-80">
        <Card className="p-6 space-y-4">
          <div>
            <h2 className="text-sm font-semibold">Recent Activity</h2>
            <p className="text-xs text-muted-foreground">
              Your latest actions and contributions
            </p>
          </div>
          <div className="space-y-4">
            {recentActivity.map((item) => (
              <div key={item.title} className="flex gap-3">
                <div className="mt-1 h-2 w-2 rounded-full" style={{ backgroundColor: 'currentColor' }}>
                  {/* small colored dot; we use tailwind class via wrapping span below */}
                </div>
                <div className="space-y-1">
                  <p className="text-sm font-medium">{item.title}</p>
                  <p className="text-xs text-muted-foreground">
                    {item.description}
                  </p>
                  <p className="text-[11px] text-muted-foreground">
                    {item.time}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </div>
  )
}
