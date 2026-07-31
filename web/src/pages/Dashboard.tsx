import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Activity, CheckCircle, XCircle, Clock, TrendingUp } from 'lucide-react';
import { mockSyncRuns, getAllMockOperations } from '@/lib/mockData';
import { formatRelativeTime, formatDuration, getStatusIcon, calculateSuccessRate } from '@/lib/utils';
import { Link } from 'react-router-dom';

export function Dashboard() {
  const allOperations = getAllMockOperations();
  const recentSyncRuns = mockSyncRuns.slice(0, 5);
  const recentOperations = allOperations.slice(0, 10);

  // Calculate stats
  const totalOperations = allOperations.length;
  const successfulOperations = allOperations.filter((op) => op.status === 'success').length;
  const failedOperations = allOperations.filter((op) => op.status === 'failed').length;
  const successRate = calculateSuccessRate(successfulOperations, totalOperations);

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold text-slate-900">Dashboard</h2>
        <p className="mt-2 text-slate-600">
          Overview of GateKeep sync operations and role management
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm font-medium">
              <Activity className="h-4 w-4 text-slate-600" />
              Total Operations
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-900">{totalOperations}</div>
            <p className="mt-1 text-xs text-slate-600">Across {mockSyncRuns.length} sync runs</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm font-medium">
              <CheckCircle className="h-4 w-4 text-green-600" />
              Successful
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{successfulOperations}</div>
            <p className="mt-1 text-xs text-slate-600">{successRate}% success rate</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm font-medium">
              <XCircle className="h-4 w-4 text-red-600" />
              Failed
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{failedOperations}</div>
            <p className="mt-1 text-xs text-slate-600">
              {((failedOperations / totalOperations) * 100).toFixed(1)}% failure rate
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm font-medium">
              <TrendingUp className="h-4 w-4 text-blue-600" />
              Recent Syncs
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-slate-900">{mockSyncRuns.length}</div>
            <p className="mt-1 text-xs text-slate-600">Last 30 days</p>
          </CardContent>
        </Card>
      </div>

      {/* Recent Sync Runs */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Recent Sync Runs</CardTitle>
            <Link
              to="/logs"
              className="text-sm text-slate-600 hover:text-slate-900 transition-colors"
            >
              View all →
            </Link>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {recentSyncRuns.map((sync) => (
              <div
                key={sync.id}
                className="flex items-center justify-between rounded-md border border-slate-200 p-3 hover:bg-slate-50 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div className="text-2xl">{getStatusIcon(sync.status)}</div>
                  <div>
                    <p className="font-mono text-sm font-medium text-slate-900">
                      Sync #{sync.id}
                    </p>
                    <p className="text-xs text-slate-500">
                      {sync.total_operations} operations • {formatRelativeTime(sync.started_at)}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <Badge
                    variant={
                      sync.status === 'success'
                        ? 'success'
                        : sync.status === 'failed'
                        ? 'error'
                        : 'warning'
                    }
                  >
                    {sync.status}
                  </Badge>
                  <span className="text-sm text-slate-500">
                    {formatDuration(sync.duration_ms)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {recentOperations.map((operation) => (
              <div
                key={operation.id}
                className="flex items-start gap-3 rounded-md p-2 hover:bg-slate-50 transition-colors"
              >
                <div className="mt-1">{getStatusIcon(operation.status)}</div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-sm font-medium text-slate-900">
                      {operation.operation_type}
                    </span>
                    <span className="text-xs text-slate-500">•</span>
                    <span className="font-mono text-xs text-slate-600">
                      {operation.target_object}
                    </span>
                  </div>
                  <p className="mt-1 text-xs text-slate-500 truncate">
                    {operation.sql_statement}
                  </p>
                  <p className="mt-1 text-xs text-slate-400">
                    {formatRelativeTime(operation.executed_at)} •{' '}
                    {formatDuration(operation.execution_time_ms)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Quick Links */}
      <div className="grid grid-cols-3 gap-4">
        <Link to="/logs">
          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <CardContent className="flex items-center gap-4 p-6">
              <div className="rounded-lg bg-blue-50 p-3">
                <Clock className="h-6 w-6 text-blue-600" />
              </div>
              <div>
                <h3 className="font-semibold text-slate-900">View Logs</h3>
                <p className="text-sm text-slate-500">Browse all operations</p>
              </div>
            </CardContent>
          </Card>
        </Link>

        <Link to="/roles">
          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <CardContent className="flex items-center gap-4 p-6">
              <div className="rounded-lg bg-purple-50 p-3">
                <Activity className="h-6 w-6 text-purple-600" />
              </div>
              <div>
                <h3 className="font-semibold text-slate-900">Role Hierarchy</h3>
                <p className="text-sm text-slate-500">Explore role relationships</p>
              </div>
            </CardContent>
          </Card>
        </Link>

        <Link to="/diff">
          <Card className="hover:shadow-md transition-shadow cursor-pointer">
            <CardContent className="flex items-center gap-4 p-6">
              <div className="rounded-lg bg-green-50 p-3">
                <CheckCircle className="h-6 w-6 text-green-600" />
              </div>
              <div>
                <h3 className="font-semibold text-slate-900">Compare Permissions</h3>
                <p className="text-sm text-slate-500">Diff tool for roles</p>
              </div>
            </CardContent>
          </Card>
        </Link>
      </div>
    </div>
  );
}
