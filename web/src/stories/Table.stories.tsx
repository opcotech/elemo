import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  ArrowUpDown,
  ChevronDown,
  ChevronUp,
  Download,
  Edit,
  Eye,
  Filter,
  MapPin,
  MoreHorizontal,
  Search,
  Trash2,
} from "lucide-react";
import { useState } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const meta: Meta<typeof Table> = {
  title: "UI/Table",
  component: Table,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A responsive table component for displaying tabular data. Built with semantic HTML table elements and enhanced with modern styling and interactive features.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    className: {
      control: "text",
      description: "Additional CSS classes to apply to the table",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic table
export const Default: Story = {
  render: () => (
    <Table>
      <TableCaption>A list of your recent invoices.</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">Invoice</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Method</TableHead>
          <TableHead className="text-right">Amount</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell className="font-medium">INV001</TableCell>
          <TableCell>Paid</TableCell>
          <TableCell>Credit Card</TableCell>
          <TableCell className="text-right">$250.00</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">INV002</TableCell>
          <TableCell>Pending</TableCell>
          <TableCell>PayPal</TableCell>
          <TableCell className="text-right">$150.00</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">INV003</TableCell>
          <TableCell>Unpaid</TableCell>
          <TableCell>Bank Transfer</TableCell>
          <TableCell className="text-right">$350.00</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">INV004</TableCell>
          <TableCell>Paid</TableCell>
          <TableCell>Credit Card</TableCell>
          <TableCell className="text-right">$450.00</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">INV005</TableCell>
          <TableCell>Paid</TableCell>
          <TableCell>PayPal</TableCell>
          <TableCell className="text-right">$550.00</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">INV006</TableCell>
          <TableCell>Pending</TableCell>
          <TableCell>Bank Transfer</TableCell>
          <TableCell className="text-right">$200.00</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">INV007</TableCell>
          <TableCell>Unpaid</TableCell>
          <TableCell>Credit Card</TableCell>
          <TableCell className="text-right">$300.00</TableCell>
        </TableRow>
      </TableBody>
      <TableFooter>
        <TableRow>
          <TableCell colSpan={3}>Total</TableCell>
          <TableCell className="text-right">$2,500.00</TableCell>
        </TableRow>
      </TableFooter>
    </Table>
  ),
};

// With status badges
export const WithStatusBadges: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Order ID</TableHead>
          <TableHead>Customer</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Date</TableHead>
          <TableHead className="text-right">Amount</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell className="font-medium">#12345</TableCell>
          <TableCell>John Doe</TableCell>
          <TableCell>
            <Badge variant="default">Completed</Badge>
          </TableCell>
          <TableCell>2023-12-01</TableCell>
          <TableCell className="text-right">$129.99</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">#12346</TableCell>
          <TableCell>Jane Smith</TableCell>
          <TableCell>
            <Badge variant="secondary">Processing</Badge>
          </TableCell>
          <TableCell>2023-12-02</TableCell>
          <TableCell className="text-right">$89.50</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">#12347</TableCell>
          <TableCell>Bob Johnson</TableCell>
          <TableCell>
            <Badge variant="destructive">Cancelled</Badge>
          </TableCell>
          <TableCell>2023-12-03</TableCell>
          <TableCell className="text-right">$45.00</TableCell>
        </TableRow>
        <TableRow>
          <TableCell className="font-medium">#12348</TableCell>
          <TableCell>Alice Brown</TableCell>
          <TableCell>
            <Badge variant="outline">Pending</Badge>
          </TableCell>
          <TableCell>2023-12-04</TableCell>
          <TableCell className="text-right">$199.99</TableCell>
        </TableRow>
      </TableBody>
    </Table>
  ),
};

// User management table
export const UserManagement: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[50px]">
            <Checkbox />
          </TableHead>
          <TableHead>User</TableHead>
          <TableHead>Role</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Last Login</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell>
            <Checkbox />
          </TableCell>
          <TableCell>
            <div className="flex items-center space-x-3">
              <Avatar className="h-8 w-8">
                <AvatarImage src="https://github.com/shadcn.png" alt="John" />
                <AvatarFallback>JD</AvatarFallback>
              </Avatar>
              <div>
                <div className="font-medium">John Doe</div>
                <div className="text-muted-foreground text-sm">
                  john@example.com
                </div>
              </div>
            </div>
          </TableCell>
          <TableCell>
            <Badge variant="default">Admin</Badge>
          </TableCell>
          <TableCell>
            <div className="flex items-center space-x-2">
              <div className="h-2 w-2 rounded-full bg-green-500"></div>
              <span>Active</span>
            </div>
          </TableCell>
          <TableCell>2 hours ago</TableCell>
          <TableCell className="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem>
                  <Eye className="size-4" />
                  View
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <Edit className="size-4" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem className="text-red-600">
                  <Trash2 className="size-4" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
        <TableRow>
          <TableCell>
            <Checkbox />
          </TableCell>
          <TableCell>
            <div className="flex items-center space-x-3">
              <Avatar className="h-8 w-8">
                <AvatarFallback>JS</AvatarFallback>
              </Avatar>
              <div>
                <div className="font-medium">Jane Smith</div>
                <div className="text-muted-foreground text-sm">
                  jane@example.com
                </div>
              </div>
            </div>
          </TableCell>
          <TableCell>
            <Badge variant="secondary">Editor</Badge>
          </TableCell>
          <TableCell>
            <div className="flex items-center space-x-2">
              <div className="h-2 w-2 rounded-full bg-green-500"></div>
              <span>Active</span>
            </div>
          </TableCell>
          <TableCell>1 day ago</TableCell>
          <TableCell className="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem>
                  <Eye className="size-4" />
                  View
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <Edit className="size-4" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem className="text-red-600">
                  <Trash2 className="size-4" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
        <TableRow>
          <TableCell>
            <Checkbox />
          </TableCell>
          <TableCell>
            <div className="flex items-center space-x-3">
              <Avatar className="h-8 w-8">
                <AvatarFallback>BJ</AvatarFallback>
              </Avatar>
              <div>
                <div className="font-medium">Bob Johnson</div>
                <div className="text-muted-foreground text-sm">
                  bob@example.com
                </div>
              </div>
            </div>
          </TableCell>
          <TableCell>
            <Badge variant="outline">Viewer</Badge>
          </TableCell>
          <TableCell>
            <div className="flex items-center space-x-2">
              <div className="h-2 w-2 rounded-full bg-gray-400"></div>
              <span>Inactive</span>
            </div>
          </TableCell>
          <TableCell>1 week ago</TableCell>
          <TableCell className="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem>
                  <Eye className="size-4" />
                  View
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <Edit className="size-4" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem className="text-red-600">
                  <Trash2 className="size-4" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  ),
};

// Sortable table
export const SortableTable: Story = {
  render: () => {
    const [sortField, setSortField] = useState<string | null>(null);
    const [sortDirection, setSortDirection] = useState<"asc" | "desc">("asc");

    const handleSort = (field: string) => {
      if (sortField === field) {
        setSortDirection(sortDirection === "asc" ? "desc" : "asc");
      } else {
        setSortField(field);
        setSortDirection("asc");
      }
    };

    const SortableHeader = ({
      field,
      children,
    }: {
      field: string;
      children: React.ReactNode;
    }) => (
      <TableHead>
        <Button
          variant="ghost"
          onClick={() => handleSort(field)}
          className="h-auto p-0 font-medium hover:bg-transparent"
        >
          {children}
          {sortField === field ? (
            sortDirection === "asc" ? (
              <ChevronUp className="ml-2 h-4 w-4" />
            ) : (
              <ChevronDown className="ml-2 h-4 w-4" />
            )
          ) : (
            <ArrowUpDown className="ml-2 h-4 w-4" />
          )}
        </Button>
      </TableHead>
    );

    return (
      <Table>
        <TableHeader>
          <TableRow>
            <SortableHeader field="name">Name</SortableHeader>
            <SortableHeader field="email">Email</SortableHeader>
            <SortableHeader field="role">Role</SortableHeader>
            <SortableHeader field="joined">Joined</SortableHeader>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow>
            <TableCell className="font-medium">Alice Wilson</TableCell>
            <TableCell>alice@company.com</TableCell>
            <TableCell>Developer</TableCell>
            <TableCell>2023-01-15</TableCell>
            <TableCell className="text-right">
              <Button variant="ghost" size="sm">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </TableCell>
          </TableRow>
          <TableRow>
            <TableCell className="font-medium">Bob Smith</TableCell>
            <TableCell>bob@company.com</TableCell>
            <TableCell>Designer</TableCell>
            <TableCell>2023-02-20</TableCell>
            <TableCell className="text-right">
              <Button variant="ghost" size="sm">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </TableCell>
          </TableRow>
          <TableRow>
            <TableCell className="font-medium">Charlie Brown</TableCell>
            <TableCell>charlie@company.com</TableCell>
            <TableCell>Manager</TableCell>
            <TableCell>2022-12-10</TableCell>
            <TableCell className="text-right">
              <Button variant="ghost" size="sm">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "A table with sortable columns that show sort indicators and handle click events.",
      },
    },
  },
};

// Data table with search and filter
export const DataTableWithSearch: Story = {
  render: () => {
    const [searchTerm, setSearchTerm] = useState("");

    return (
      <div className="space-y-4">
        <div className="flex items-center space-x-2">
          <div className="relative flex-1">
            <Search className="text-muted-foreground absolute top-2.5 left-2 h-4 w-4" />
            <Input
              placeholder="Search customers..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-8"
            />
          </div>
          <Button variant="outline">
            <Filter className="size-4" />
            Filter
          </Button>
          <Button variant="outline">
            <Download className="size-4" />
            Export
          </Button>
        </div>

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[50px]">
                <Checkbox />
              </TableHead>
              <TableHead>Customer</TableHead>
              <TableHead>Company</TableHead>
              <TableHead>Phone</TableHead>
              <TableHead>Location</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>
                <Checkbox />
              </TableCell>
              <TableCell>
                <div className="flex items-center space-x-3">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>AM</AvatarFallback>
                  </Avatar>
                  <div>
                    <div className="font-medium">Anna Martinez</div>
                    <div className="text-muted-foreground text-sm">
                      anna@techcorp.com
                    </div>
                  </div>
                </div>
              </TableCell>
              <TableCell>TechCorp Inc.</TableCell>
              <TableCell>+1 (555) 123-4567</TableCell>
              <TableCell>
                <div className="flex items-center space-x-1">
                  <MapPin className="h-3 w-3" />
                  <span>New York, NY</span>
                </div>
              </TableCell>
              <TableCell className="text-right">
                <div className="flex items-center justify-end space-x-1">
                  <Button variant="ghost" size="sm">
                    <Eye className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="sm">
                    <Edit className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="sm">
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>
                <Checkbox />
              </TableCell>
              <TableCell>
                <div className="flex items-center space-x-3">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>DJ</AvatarFallback>
                  </Avatar>
                  <div>
                    <div className="font-medium">David Johnson</div>
                    <div className="text-muted-foreground text-sm">
                      david@startup.io
                    </div>
                  </div>
                </div>
              </TableCell>
              <TableCell>Startup Labs</TableCell>
              <TableCell>+1 (555) 987-6543</TableCell>
              <TableCell>
                <div className="flex items-center space-x-1">
                  <MapPin className="h-3 w-3" />
                  <span>San Francisco, CA</span>
                </div>
              </TableCell>
              <TableCell className="text-right">
                <div className="flex items-center justify-end space-x-1">
                  <Button variant="ghost" size="sm">
                    <Eye className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="sm">
                    <Edit className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="sm">
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>
                <Checkbox />
              </TableCell>
              <TableCell>
                <div className="flex items-center space-x-3">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>EB</AvatarFallback>
                  </Avatar>
                  <div>
                    <div className="font-medium">Emma Brown</div>
                    <div className="text-muted-foreground text-sm">
                      emma@enterprise.com
                    </div>
                  </div>
                </div>
              </TableCell>
              <TableCell>Enterprise Solutions</TableCell>
              <TableCell>+1 (555) 246-8135</TableCell>
              <TableCell>
                <div className="flex items-center space-x-1">
                  <MapPin className="h-3 w-3" />
                  <span>Chicago, IL</span>
                </div>
              </TableCell>
              <TableCell className="text-right">
                <div className="flex items-center justify-end space-x-1">
                  <Button variant="ghost" size="sm">
                    <Eye className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="sm">
                    <Edit className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="sm">
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story:
          "A data table with search functionality, filters, and bulk actions.",
      },
    },
  },
};

// Product inventory table
export const ProductInventory: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Product</TableHead>
          <TableHead>SKU</TableHead>
          <TableHead>Category</TableHead>
          <TableHead>Stock</TableHead>
          <TableHead>Price</TableHead>
          <TableHead>Status</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell>
            <div className="flex items-center space-x-3">
              <div className="bg-muted flex h-10 w-10 items-center justify-center rounded">
                📱
              </div>
              <div>
                <div className="font-medium">iPhone 15 Pro</div>
                <div className="text-muted-foreground text-sm">Apple</div>
              </div>
            </div>
          </TableCell>
          <TableCell className="font-mono">IP15P-256-TB</TableCell>
          <TableCell>Electronics</TableCell>
          <TableCell>
            <div className="flex items-center space-x-2">
              <div className="h-2 w-2 rounded-full bg-green-500"></div>
              <span>42 units</span>
            </div>
          </TableCell>
          <TableCell>$999.00</TableCell>
          <TableCell>
            <Badge variant="default">In Stock</Badge>
          </TableCell>
          <TableCell className="text-right">
            <Button variant="ghost" size="sm">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </TableCell>
        </TableRow>
        <TableRow>
          <TableCell>
            <div className="flex items-center space-x-3">
              <div className="bg-muted flex h-10 w-10 items-center justify-center rounded">
                💻
              </div>
              <div>
                <div className="font-medium">MacBook Air M3</div>
                <div className="text-muted-foreground text-sm">Apple</div>
              </div>
            </div>
          </TableCell>
          <TableCell className="font-mono">MBA-M3-512-SG</TableCell>
          <TableCell>Computers</TableCell>
          <TableCell>
            <div className="flex items-center space-x-2">
              <div className="h-2 w-2 rounded-full bg-yellow-500"></div>
              <span>8 units</span>
            </div>
          </TableCell>
          <TableCell>$1,299.00</TableCell>
          <TableCell>
            <Badge variant="secondary">Low Stock</Badge>
          </TableCell>
          <TableCell className="text-right">
            <Button variant="ghost" size="sm">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </TableCell>
        </TableRow>
        <TableRow>
          <TableCell>
            <div className="flex items-center space-x-3">
              <div className="bg-muted flex h-10 w-10 items-center justify-center rounded">
                🎧
              </div>
              <div>
                <div className="font-medium">AirPods Pro</div>
                <div className="text-muted-foreground text-sm">Apple</div>
              </div>
            </div>
          </TableCell>
          <TableCell className="font-mono">APP-2ND-GEN</TableCell>
          <TableCell>Audio</TableCell>
          <TableCell>
            <div className="flex items-center space-x-2">
              <div className="h-2 w-2 rounded-full bg-red-500"></div>
              <span>0 units</span>
            </div>
          </TableCell>
          <TableCell>$249.00</TableCell>
          <TableCell>
            <Badge variant="destructive">Out of Stock</Badge>
          </TableCell>
          <TableCell className="text-right">
            <Button variant="ghost" size="sm">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  ),
};

// Minimal table
export const MinimalTable: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Email</TableHead>
          <TableHead>Role</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell>John Doe</TableCell>
          <TableCell>john@example.com</TableCell>
          <TableCell>Developer</TableCell>
        </TableRow>
        <TableRow>
          <TableCell>Jane Smith</TableCell>
          <TableCell>jane@example.com</TableCell>
          <TableCell>Designer</TableCell>
        </TableRow>
        <TableRow>
          <TableCell>Bob Johnson</TableCell>
          <TableCell>bob@example.com</TableCell>
          <TableCell>Manager</TableCell>
        </TableRow>
      </TableBody>
    </Table>
  ),
};

// Dense table
export const DenseTable: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow className="h-8">
          <TableHead className="h-8 text-xs">ID</TableHead>
          <TableHead className="h-8 text-xs">Name</TableHead>
          <TableHead className="h-8 text-xs">Status</TableHead>
          <TableHead className="h-8 text-xs">Date</TableHead>
          <TableHead className="h-8 text-right text-xs">Amount</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 10 }, (_, i) => (
          <TableRow key={i} className="h-8">
            <TableCell className="h-8 font-mono text-xs">
              {String(i + 1).padStart(3, "0")}
            </TableCell>
            <TableCell className="h-8 text-xs">Item {i + 1}</TableCell>
            <TableCell className="h-8 text-xs">
              <Badge
                variant={i % 3 === 0 ? "default" : "secondary"}
                className="h-5 text-xs"
              >
                {i % 3 === 0 ? "Active" : "Pending"}
              </Badge>
            </TableCell>
            <TableCell className="h-8 text-xs">
              2023-12-{String(i + 1).padStart(2, "0")}
            </TableCell>
            <TableCell className="h-8 text-right text-xs">
              ${(i + 1) * 10}.00
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "A compact table with reduced padding and smaller text for displaying more data.",
      },
    },
  },
};
