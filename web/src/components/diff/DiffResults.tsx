import { useState } from 'react';
import { Input } from '@/components/ui/input';
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
  const [objectFilter, setObjectFilter] = useState('');
  const [objectType, setObjectType] = useState('');
  const allGrants = [...added, ...removed, ...unchanged];
  const types = [...new Set(allGrants.map(grant => grant.object_type))].sort();
  const matches = (grant: Grant) => (!objectType || grant.object_type === objectType) && grant.object_name.toLowerCase().includes(objectFilter.trim().toLowerCase());
  const filteredAdded = added.filter(matches);
  const filteredRemoved = removed.filter(matches);
  const filteredUnchanged = unchanged.filter(matches);
  const total = filteredAdded.length + filteredRemoved.length + filteredUnchanged.length;

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
      <div className="flex flex-wrap items-end gap-4">
        <label className="min-w-[240px] flex-1 text-sm font-medium text-slate-700">Object name
          <Input aria-label="Filter by object name" placeholder="Search database, schema, or object…" value={objectFilter} onChange={event => setObjectFilter(event.target.value)} className="mt-1" />
        </label>
        <label className="text-sm font-medium text-slate-700">Object type
          <select aria-label="Filter by object type" value={objectType} onChange={event => setObjectType(event.target.value)} className="mt-1 block rounded-md border border-slate-300 bg-white p-2 text-sm text-slate-900">
            <option value="">All object types</option>
            {types.map(type => <option key={type} value={type}>{type}</option>)}
          </select>
        </label>
      </div>
      <p className="text-sm text-slate-500">Showing {total} of {allGrants.length} permissions</p>
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <Plus className="h-4 w-4 text-green-600" />
              Added
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-green-600">{filteredAdded.length}</div>
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
            <div className="text-3xl font-bold text-red-600">{filteredRemoved.length}</div>
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
            <div className="text-3xl font-bold text-slate-600">{filteredUnchanged.length}</div>
            <p className="text-xs text-slate-500 mt-1">
              Common permissions
            </p>
          </CardContent>
        </Card>
      </div>

      {total === 0 && <p>No permissions match the selected object filters.</p>}
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
                {filteredAdded.map((grant, index) => (
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
                {filteredRemoved.map((grant, index) => (
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
                {filteredUnchanged.map((grant, index) => (
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

              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
