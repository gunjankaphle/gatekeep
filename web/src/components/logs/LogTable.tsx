import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import type { Operation } from '@/types';
import {
  formatRelativeTime,
  formatDuration,
  getStatusIcon,
  truncate,
} from '@/lib/utils';

interface LogTableProps {
  operations: Operation[];
  onRowClick: (operation: Operation) => void;
}

export function LogTable({ operations, onRowClick }: LogTableProps) {
  const getStatusVariant = (status: string) => {
    switch (status) {
      case 'success':
        return 'success';
      case 'failed':
        return 'error';
      case 'skipped':
        return 'outline';
      default:
        return 'default';
    }
  };

  if (operations.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center rounded-lg border border-slate-200 bg-white">
        <div className="text-center">
          <p className="text-sm text-slate-500">No operations found</p>
          <p className="mt-1 text-xs text-slate-400">Try adjusting your filters</p>
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-slate-200 bg-white">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Timestamp</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Operation Type</TableHead>
            <TableHead>Target Object</TableHead>
            <TableHead>SQL Statement</TableHead>
            <TableHead className="text-right">Duration</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {operations.map((operation) => (
            <TableRow
              key={operation.id}
              className="cursor-pointer"
              onClick={() => onRowClick(operation)}
            >
              <TableCell className="font-mono text-xs text-slate-500">
                {formatRelativeTime(operation.executed_at)}
              </TableCell>
              <TableCell>
                <Badge variant={getStatusVariant(operation.status)}>
                  {getStatusIcon(operation.status)} {operation.status}
                </Badge>
              </TableCell>
              <TableCell>
                <span className="font-mono text-sm">{operation.operation_type}</span>
              </TableCell>
              <TableCell>
                <span className="font-mono text-sm">{operation.target_object}</span>
              </TableCell>
              <TableCell className="max-w-md">
                <code className="text-xs text-slate-600">
                  {truncate(operation.sql_statement, 80)}
                </code>
              </TableCell>
              <TableCell className="text-right font-mono text-sm text-slate-500">
                {formatDuration(operation.execution_time_ms)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
