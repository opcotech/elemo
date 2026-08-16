import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export type TableSkeletonColumn = {
  header: string;
  skeletonClassName: string;
  headerClassName?: string;
  cellClassName?: string;
  srOnly?: boolean;
  count?: number;
};

function TableSkeletonCell({ column }: { column: TableSkeletonColumn }) {
  const count = column.count ?? 1;
  const skeletons = Array.from({ length: count }).map((_, index) => (
    <Skeleton key={index} className={column.skeletonClassName} />
  ));

  return (
    <TableCell className={column.cellClassName}>
      {count > 1 ? (
        <div className="flex justify-end gap-1">{skeletons}</div>
      ) : (
        skeletons
      )}
    </TableCell>
  );
}

export function TableSkeletonRows({
  columns,
  rows = 5,
}: {
  columns: readonly TableSkeletonColumn[];
  rows?: number;
}) {
  return (
    <>
      {Array.from({ length: rows }).map((_, rowIndex) => (
        <TableRow key={rowIndex}>
          {columns.map((column) => (
            <TableSkeletonCell key={column.header} column={column} />
          ))}
        </TableRow>
      ))}
    </>
  );
}

export function TableSkeleton({
  columns,
  rows = 5,
}: {
  columns: readonly TableSkeletonColumn[];
  rows?: number;
}) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          {columns.map((column) => (
            <TableHead key={column.header} className={column.headerClassName}>
              {column.srOnly ? (
                <span className="sr-only">{column.header}</span>
              ) : (
                column.header
              )}
            </TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableSkeletonRows columns={columns} rows={rows} />
      </TableBody>
    </Table>
  );
}
