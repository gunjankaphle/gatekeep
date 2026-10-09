import { useState, useMemo } from 'react';
import { LogFilters } from './LogFilters';
import { LogTable } from './LogTable';
import { LogDetailsModal } from './LogDetailsModal';
import { Button } from '@/components/ui/button';
import type { LogFilters as LogFiltersType, Operation } from '@/types';
import { useAuditData } from '@/lib/queries';
import { QueryState } from '@/components/ui/QueryState';
import { ChevronLeft, ChevronRight } from 'lucide-react';

const ITEMS_PER_PAGE = 50;

export function LogViewer() {
  const [filters, setFilters] = useState<LogFiltersType>({});
  const [selectedOperation, setSelectedOperation] = useState<Operation | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);

  // Get all operations from mock data
  const query = useAuditData();


  // Filter operations based on active filters
  const filteredOperations = useMemo(() => {
    return (query.data?.operations ?? []).filter((op) => {
      // Status filter
      if (filters.status && op.status !== filters.status) {
        return false;
      }

      // Operation type filter
      if (filters.operation_type && op.operation_type !== filters.operation_type) {
        return false;
      }

      // Search filter (searches in target_object and sql_statement)
      if (filters.search) {
        const searchLower = filters.search.toLowerCase();
        const matchesTarget = op.target_object.toLowerCase().includes(searchLower);
        const matchesSQL = op.sql_statement.toLowerCase().includes(searchLower);
        if (!matchesTarget && !matchesSQL) {
          return false;
        }
      }

      return true;
    });
  }, [query.data, filters]);

  // Calculate pagination
  const totalPages = Math.ceil(filteredOperations.length / ITEMS_PER_PAGE);
  const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
  const endIndex = startIndex + ITEMS_PER_PAGE;
  const paginatedOperations = filteredOperations.slice(startIndex, endIndex);



  const handleRowClick = (operation: Operation) => {
    setSelectedOperation(operation);
    setModalOpen(true);
  };

  const handlePreviousPage = () => {
    setCurrentPage((prev) => Math.max(1, prev - 1));
  };

  const handleNextPage = () => {
    setCurrentPage((prev) => Math.min(totalPages, prev + 1));
  };

  const handlePageClick = (page: number) => {
    setCurrentPage(page);
  };

  // Generate page numbers to show
  const getPageNumbers = () => {
    const pages: (number | string)[] = [];
    const showPages = 5; // Show 5 page numbers at a time

    if (totalPages <= showPages) {
      // Show all pages if total is small
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      // Show first page
      pages.push(1);

      // Show pages around current page
      const start = Math.max(2, currentPage - 1);
      const end = Math.min(totalPages - 1, currentPage + 1);

      if (start > 2) pages.push('...');

      for (let i = start; i <= end; i++) {
        pages.push(i);
      }

      if (end < totalPages - 1) pages.push('...');

      // Show last page
      pages.push(totalPages);
    }

    return pages;
  };

  if (query.isPending || query.isError) return <QueryState pending={query.isPending} error={query.error} />;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold text-slate-900">Log Viewer</h2>
        <p className="mt-2 text-slate-600">
          View and filter operations from the latest 100 sync runs
        </p>
      </div>

      <div className="space-y-4">
        <div className="w-full">
          <LogFilters filters={filters} onFiltersChange={(next) => { setFilters(next); setCurrentPage(1); }} />
        </div>

        <div className="min-w-0 space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-slate-600">
              Showing{' '}
              <span className="font-semibold">
                {filteredOperations.length ? startIndex + 1 : 0}-{Math.min(endIndex, filteredOperations.length)}
              </span>{' '}
              of <span className="font-semibold">{filteredOperations.length}</span> operations
            </p>
          </div>

          <LogTable operations={paginatedOperations} onRowClick={handleRowClick} />

          {/* Pagination Controls */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between rounded-lg border border-slate-200 bg-white px-4 py-3">
              <div className="text-sm text-slate-600">
                Page <span className="font-semibold">{currentPage}</span> of{' '}
                <span className="font-semibold">{totalPages}</span>
              </div>

              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handlePreviousPage}
                  disabled={currentPage === 1}
                >
                  <ChevronLeft className="h-4 w-4" />
                  Previous
                </Button>

                <div className="flex gap-1">
                  {getPageNumbers().map((page, index) =>
                    typeof page === 'number' ? (
                      <Button
                        key={index}
                        variant={currentPage === page ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => handlePageClick(page)}
                        className="min-w-[40px]"
                      >
                        {page}
                      </Button>
                    ) : (
                      <span key={index} className="flex items-center px-2 text-slate-400">
                        {page}
                      </span>
                    )
                  )}
                </div>

                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleNextPage}
                  disabled={currentPage === totalPages}
                >
                  Next
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>

      <LogDetailsModal
        operation={selectedOperation}
        open={modalOpen}
        onOpenChange={setModalOpen}
      />
    </div>
  );
}
