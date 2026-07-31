# GateKeep Web UI

A modern, read-only web interface for GateKeep - Snowflake permission management tool.

## Features

### 1. Log Viewer
- Browse and filter 90+ mock sync operations
- Search by role name, object, or SQL statement
- Filter by status (success/failed), operation type (CREATE_ROLE, GRANT, REVOKE)
- View detailed operation information including SQL statements
- Modal view for full operation details

### 2. Role Hierarchy Explorer
- **Graph View**: Interactive DAG visualization using ReactFlow
  - Zoom, pan, and explore role relationships
  - Click nodes to view role details
  - Auto-layout with dagre algorithm
- **Tree View**: Collapsible tree structure
  - Parent-child relationships
  - Easy navigation and exploration
- **Role Details Panel**: View role metadata, parent roles, and permissions

### 3. Permission Diff Tool
- Compare privileges between two roles
- Visual diff with color-coded results:
  - ✅ Green: Added permissions (in Role B, not in Role A)
  - ❌ Red: Removed permissions (in Role A, not in Role B)
  - ⚪ Gray: Unchanged permissions (common to both)
- Summary statistics and detailed permission tables

## Technology Stack

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS + custom components
- **Routing**: React Router v6
- **DAG Visualization**: ReactFlow
- **State Management**: React hooks (TanStack Query ready)
- **Icons**: Lucide React
- **Date Handling**: date-fns

## Getting Started

### Prerequisites
- Node.js 18+ and npm

### Installation

```bash
# Install dependencies
npm install
```

### Development

```bash
# Start development server
npm run dev
```

The app will be available at `http://localhost:5173`

### Build for Production

```bash
# Build static assets
npm run build
```

Output will be in the `dist/` directory.

## Mock Data

The UI currently uses comprehensive mock data with 90+ operations across 10 sync runs.

## License

Part of the GateKeep project.
