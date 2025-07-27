import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";

const meta: Meta<typeof Pagination> = {
  title: "UI/Pagination",
  component: Pagination,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "A pagination component for navigating through pages of content. Provides accessible navigation controls with customizable styling.",
      },
    },
  },
  tags: ["autodocs"],
  argTypes: {
    className: {
      control: "text",
      description: "Additional CSS classes to apply to the pagination",
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

// Basic pagination
export const Default: Story = {
  render: () => (
    <Pagination>
      <PaginationContent>
        <PaginationItem>
          <PaginationPrevious href="#" />
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">1</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#" isActive>
            2
          </PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">3</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationEllipsis />
        </PaginationItem>
        <PaginationItem>
          <PaginationNext href="#" />
        </PaginationItem>
      </PaginationContent>
    </Pagination>
  ),
};

// Simple pagination
export const Simple: Story = {
  render: () => (
    <Pagination>
      <PaginationContent>
        <PaginationItem>
          <PaginationPrevious href="#" />
        </PaginationItem>
        <PaginationItem>
          <PaginationNext href="#" />
        </PaginationItem>
      </PaginationContent>
    </Pagination>
  ),
};

// With ellipsis
export const WithEllipsis: Story = {
  render: () => (
    <Pagination>
      <PaginationContent>
        <PaginationItem>
          <PaginationPrevious href="#" />
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">1</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationEllipsis />
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#" isActive>
            5
          </PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">6</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">7</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationEllipsis />
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">20</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationNext href="#" />
        </PaginationItem>
      </PaginationContent>
    </Pagination>
  ),
};

// Many pages
export const ManyPages: Story = {
  render: () => (
    <Pagination>
      <PaginationContent>
        <PaginationItem>
          <PaginationPrevious href="#" />
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">1</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">2</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">3</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#" isActive>
            4
          </PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">5</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">6</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink href="#">7</PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationNext href="#" />
        </PaginationItem>
      </PaginationContent>
    </Pagination>
  ),
};

// Disabled states
export const DisabledStates: Story = {
  render: () => (
    <div className="space-y-4">
      <div className="text-muted-foreground text-sm">
        First page (previous disabled):
      </div>
      <Pagination>
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious
              href="#"
              className="pointer-events-none opacity-50"
            />
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#" isActive>
              1
            </PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">2</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">3</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationNext href="#" />
          </PaginationItem>
        </PaginationContent>
      </Pagination>

      <div className="text-muted-foreground text-sm">
        Last page (next disabled):
      </div>
      <Pagination>
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious href="#" />
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">8</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">9</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#" isActive>
              10
            </PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationNext
              href="#"
              className="pointer-events-none opacity-50"
            />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story:
          "Pagination with disabled previous/next buttons at the boundaries.",
      },
    },
  },
};

// Controlled pagination
export const Controlled: Story = {
  render: () => {
    const [currentPage, setCurrentPage] = useState(3);
    const totalPages = 10;

    const handlePageChange = (page: number) => {
      setCurrentPage(page);
    };

    const renderPageNumbers = () => {
      const pages = [];
      const startPage = Math.max(1, currentPage - 2);
      const endPage = Math.min(totalPages, currentPage + 2);

      if (startPage > 1) {
        pages.push(
          <PaginationItem key={1}>
            <PaginationLink
              href="#"
              onClick={(e) => {
                e.preventDefault();
                handlePageChange(1);
              }}
            >
              1
            </PaginationLink>
          </PaginationItem>
        );
        if (startPage > 2) {
          pages.push(
            <PaginationItem key="ellipsis1">
              <PaginationEllipsis />
            </PaginationItem>
          );
        }
      }

      for (let i = startPage; i <= endPage; i++) {
        pages.push(
          <PaginationItem key={i}>
            <PaginationLink
              href="#"
              isActive={i === currentPage}
              onClick={(e) => {
                e.preventDefault();
                handlePageChange(i);
              }}
            >
              {i}
            </PaginationLink>
          </PaginationItem>
        );
      }

      if (endPage < totalPages) {
        if (endPage < totalPages - 1) {
          pages.push(
            <PaginationItem key="ellipsis2">
              <PaginationEllipsis />
            </PaginationItem>
          );
        }
        pages.push(
          <PaginationItem key={totalPages}>
            <PaginationLink
              href="#"
              onClick={(e) => {
                e.preventDefault();
                handlePageChange(totalPages);
              }}
            >
              {totalPages}
            </PaginationLink>
          </PaginationItem>
        );
      }

      return pages;
    };

    return (
      <div className="space-y-4">
        <div className="text-muted-foreground text-center text-sm">
          Page {currentPage} of {totalPages}
        </div>
        <Pagination>
          <PaginationContent>
            <PaginationItem>
              <PaginationPrevious
                href="#"
                onClick={(e) => {
                  e.preventDefault();
                  if (currentPage > 1) handlePageChange(currentPage - 1);
                }}
                className={
                  currentPage === 1 ? "pointer-events-none opacity-50" : ""
                }
              />
            </PaginationItem>
            {renderPageNumbers()}
            <PaginationItem>
              <PaginationNext
                href="#"
                onClick={(e) => {
                  e.preventDefault();
                  if (currentPage < totalPages)
                    handlePageChange(currentPage + 1);
                }}
                className={
                  currentPage === totalPages
                    ? "pointer-events-none opacity-50"
                    : ""
                }
              />
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      </div>
    );
  },
  parameters: {
    docs: {
      description: {
        story: "A fully controlled pagination component with state management.",
      },
    },
  },
};

// Custom styling
export const CustomStyling: Story = {
  render: () => (
    <Pagination className="bg-muted/50 rounded-lg p-4">
      <PaginationContent>
        <PaginationItem>
          <PaginationPrevious
            href="#"
            className="bg-background hover:bg-primary/5 hover:text-primary hover:border-primary/20 border border-transparent transition-colors"
          />
        </PaginationItem>
        <PaginationItem>
          <PaginationLink
            href="#"
            className="bg-background hover:bg-primary/5 hover:text-primary hover:border-primary/20 border border-transparent transition-colors"
          >
            1
          </PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink
            href="#"
            isActive
            className="bg-primary text-primary-foreground border-primary"
          >
            2
          </PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationLink
            href="#"
            className="bg-background hover:bg-primary/5 hover:text-primary hover:border-primary/20 border border-transparent transition-colors"
          >
            3
          </PaginationLink>
        </PaginationItem>
        <PaginationItem>
          <PaginationEllipsis className="text-muted-foreground" />
        </PaginationItem>
        <PaginationItem>
          <PaginationNext
            href="#"
            className="bg-background hover:bg-primary/5 hover:text-primary hover:border-primary/20 border border-transparent transition-colors"
          />
        </PaginationItem>
      </PaginationContent>
    </Pagination>
  ),
  parameters: {
    docs: {
      description: {
        story: "Pagination with custom styling and background.",
      },
    },
  },
};

// With page info
export const WithPageInfo: Story = {
  render: () => (
    <div className="space-y-4">
      <div className="text-center">
        <div className="text-muted-foreground text-sm">
          Showing 21-30 of 97 results
        </div>
      </div>
      <Pagination>
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious href="#" />
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">2</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#" isActive>
              3
            </PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">4</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationEllipsis />
          </PaginationItem>
          <PaginationItem>
            <PaginationLink href="#">10</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationNext href="#" />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: "Pagination with additional page information display.",
      },
    },
  },
};

// Compact pagination
export const Compact: Story = {
  render: () => (
    <Pagination>
      <PaginationContent>
        <PaginationItem>
          <Button variant="outline" size="sm">
            <ChevronsLeft className="h-4 w-4" />
          </Button>
        </PaginationItem>
        <PaginationItem>
          <Button variant="outline" size="sm">
            <ChevronLeft className="h-4 w-4" />
          </Button>
        </PaginationItem>
        <PaginationItem>
          <div className="flex items-center px-3 py-1 text-sm">
            Page 3 of 10
          </div>
        </PaginationItem>
        <PaginationItem>
          <Button variant="outline" size="sm">
            <ChevronRight className="h-4 w-4" />
          </Button>
        </PaginationItem>
        <PaginationItem>
          <Button variant="outline" size="sm">
            <ChevronsRight className="h-4 w-4" />
          </Button>
        </PaginationItem>
      </PaginationContent>
    </Pagination>
  ),
  parameters: {
    docs: {
      description: {
        story: "A compact pagination variant with first/last page navigation.",
      },
    },
  },
};
