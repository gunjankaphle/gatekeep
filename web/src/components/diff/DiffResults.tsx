import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Plus, Minus, Equal } from 'lucide-react';
import type { Grant } from '@/types';

interface DiffResultsProps {
  added: Grant[];
  removed: Grant[];
  unchanged: Grant[];
  roleA: string;
  roleB: string;
}

export function DiffResults({ added, removed, unchanged, roleA, roleB }: DiffResultsProps) {
  const total = added.length + removed.length + unchanged.length;

  if (!roleA || !roleB) {
    return (
      <Card>
        <CardContent className="flex h-96 items-center justify-center">
          <p className="text-sm text-slate-500">
            Select two roles to compare their permissions
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <Plus className="h-4 w-4 text-green-600" />
              Added
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-green-600">{added.length}</div>
            <p className="text-xs text-slate-500 mt-1">
              Permissions in {roleB} not in {roleA}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <Minus className="h-4 w-4 text-red-600" />
              Removed
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-red-600">{removed.length}</div>
            <p className="text-xs text-slate-500 mt-1">
              Permissions in {roleA} not in {roleB}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <Equal className="h-4 w-4 text-slate-600" />
              Unchanged
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-slate-600">{unchanged.length}</div>
            <p className="text-xs text-slate-500 mt-1">
              Common permissions
            </p>
          </CardContent>
        </Card>
      </div>

      {total > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Permission Details</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Status</TableHead>
                  <TableHead>Object Type</TableHead>
                  <TableHead>Object Name</TableHead>
                  <TableHead>Privilege</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {added.map((grant, index) => (
                  <TableRow key={`added-${index}`}>
                    <TableCell>
                      <Badge variant="success">
                        <Plus className="mr-1 h-3 w-3" />
                        Added
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm text-slate-900">
                      {grant.object_type}
                    </TableCell>
                    <TableCell className="font-mono text-sm text-slate-900">
                      {grant.object_name}
                    </TableCell>
                    <TableCell className="font-mono text-sm text-slate-900">
                      {grant.privilege}
                    </TableCell>
                  </TableRow>
                ))}
                {removed.map((grant, index) => (
                  <TableRow key={`removed-${index}`}>
                    <TableCell>
                      <Badge variant="error">
                        <Minus className="mr-1 h-3 w-3" />
                        Removed
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm text-slate-900">
                      {grant.object_type}
                    </TableCell>
                    <TableCell className="font-mono text-sm text-slate-900">
                      {grant.object_name}
                    </TableCell>
                    <TableCell className="font-mono text-sm text-slate-900">
                      {grant.privilege}
                    </TableCell>
                  </TableRow>
                ))}
                {unchanged.slice(0, 5).map((grant, index) => (
                  <TableRow key={`unchanged-${index}`}>
                    <TableCell>
                      <Badge variant="outline">
                        <Equal className="mr-1 h-3 w-3" />
                        Unchanged
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm text-slate-900">
                      {grant.object_type}
                    </TableCell>
                    <TableCell className="font-mono text-sm text-slate-900">
                      {grant.object_name}
                    </TableCell>
                    <TableCell className="font-mono text-sm text-slate-900">
                      {grant.privilege}
                    </TableCell>
                  </TableRow>
                ))}
                {unchanged.length > 5 && (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center text-sm text-slate-500">
                      ... and {unchanged.length - 5} more unchanged permissions
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
