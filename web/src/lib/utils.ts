import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { format, formatDistanceToNow } from 'date-fns';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// Format duration from milliseconds to human-readable string
export function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`;
  }
  if (ms < 60000) {
    return `${(ms / 1000).toFixed(2)}s`;
  }
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

// Format timestamp to relative time (e.g., "2 hours ago")
export function formatRelativeTime(timestamp: string): string {
  return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
}

// Format timestamp to absolute time
export function formatAbsoluteTime(timestamp: string): string {
  return format(new Date(timestamp), 'PPpp');
}

// Calculate success rate percentage
export function calculateSuccessRate(successful: number, total: number): number {
  if (total === 0) return 0;
  return Math.round((successful / total) * 100);
}

// Get status color class
export function getStatusColor(status: 'success' | 'failed' | 'partial' | 'pending' | 'running' | 'skipped'): string {
  switch (status) {
    case 'success':
      return 'text-green-600 bg-green-50 border-green-200';
    case 'failed':
      return 'text-red-600 bg-red-50 border-red-200';
    case 'partial':
      return 'text-yellow-600 bg-yellow-50 border-yellow-200';
    case 'pending':
      return 'text-gray-600 bg-gray-50 border-gray-200';
    case 'running':
      return 'text-blue-600 bg-blue-50 border-blue-200';
    case 'skipped':
      return 'text-gray-400 bg-gray-50 border-gray-200';
    default:
      return 'text-gray-600 bg-gray-50 border-gray-200';
  }
}

// Get status icon
export function getStatusIcon(status: 'success' | 'failed' | 'partial' | 'pending' | 'running' | 'skipped'): string {
  switch (status) {
    case 'success':
      return '✅';
    case 'failed':
      return '❌';
    case 'partial':
      return '⚠️';
    case 'pending':
      return '⏳';
    case 'running':
      return '▶️';
    case 'skipped':
      return '⏭️';
    default:
      return '❓';
  }
}

// Truncate text with ellipsis
export function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

// Extract role name from SQL statement
export function extractRoleFromSQL(sql: string): string | null {
  const patterns = [
    /CREATE ROLE (\w+)/i,
    /GRANT .+ TO ROLE (\w+)/i,
    /REVOKE .+ FROM ROLE (\w+)/i,
    /ROLE (\w+)/i,
  ];

  for (const pattern of patterns) {
    const match = sql.match(pattern);
    if (match) return match[1];
  }
  return null;
}
