import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogClose,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Copy, Check } from 'lucide-react';
import type { Operation } from '@/types';
import { formatDuration, formatAbsoluteTime, getStatusIcon } from '@/lib/utils';
import { useState } from 'react';

interface LogDetailsModalProps {
  operation: Operation | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function LogDetailsModal({ operation, open, onOpenChange }: LogDetailsModalProps) {
  const [copied, setCopied] = useState(false);

  if (!operation) return null;

  const handleCopySQL = () => {
    navigator.clipboard.writeText(operation.sql_statement);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

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

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogClose onClick={() => onOpenChange(false)} />
        <DialogHeader>
          <DialogTitle>Operation Details</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium text-slate-500">Operation Type</label>
              <p className="mt-1 font-mono text-sm">{operation.operation_type}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-slate-500">Target Object</label>
              <p className="mt-1 font-mono text-sm">{operation.target_object}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-slate-500">Status</label>
              <div className="mt-1">
                <Badge variant={getStatusVariant(operation.status)}>
                  {getStatusIcon(operation.status)} {operation.status.toUpperCase()}
                </Badge>
              </div>
            </div>
            <div>
              <label className="text-sm font-medium text-slate-500">Execution Time</label>
              <p className="mt-1 text-sm">{formatDuration(operation.execution_time_ms)}</p>
            </div>
            <div className="col-span-2">
              <label className="text-sm font-medium text-slate-500">Executed At</label>
              <p className="mt-1 text-sm">{formatAbsoluteTime(operation.executed_at)}</p>
            </div>
          </div>

          {operation.error_message && (
            <div className="rounded-md border border-red-200 bg-red-50 p-3">
              <label className="text-sm font-medium text-red-900">Error Message</label>
              <p className="mt-1 text-sm text-red-700">{operation.error_message}</p>
            </div>
          )}

          <div>
            <div className="mb-2 flex items-center justify-between">
              <label className="text-sm font-medium text-slate-500">SQL Statement</label>
              <Button variant="outline" size="sm" onClick={handleCopySQL}>
                {copied ? (
                  <>
                    <Check className="mr-1 h-4 w-4" />
                    Copied
                  </>
                ) : (
                  <>
                    <Copy className="mr-1 h-4 w-4" />
                    Copy
                  </>
                )}
              </Button>
            </div>
            <pre className="rounded-md border border-slate-200 bg-slate-50 p-4 text-sm overflow-x-auto">
              <code>{operation.sql_statement}</code>
            </pre>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
