import { useCallback, useMemo, useEffect, useState } from 'react';
import ReactFlow, {
  type Node,
  type Edge,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  MarkerType,
  Position,
} from 'reactflow';
import dagre from 'dagre';
import 'reactflow/dist/style.css';
import type { Role } from '@/types';
import { Button } from '@/components/ui/button';
import { RotateCcw } from 'lucide-react';

interface RoleGraphProps {
  roles: Role[];
  onRoleClick: (role: Role) => void;
}

const dagreGraph = new dagre.graphlib.Graph();
dagreGraph.setDefaultEdgeLabel(() => ({}));

const nodeWidth = 200;
const nodeHeight = 80;

const getLayoutedElements = (nodes: Node[], edges: Edge[]) => {
  dagreGraph.setGraph({ rankdir: 'TB', ranksep: 100, nodesep: 80 });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
      sourcePosition: Position.Bottom,
      targetPosition: Position.Top,
    };
  });

  return { nodes: layoutedNodes, edges };
};

// Determine node color based on hierarchy level
const getNodeColor = (role: Role, roles: Role[]): string => {
  if (role.parent_roles.length === 0) {
    return '#3b82f6'; // Blue for root
  }

  const hasChildren = roles.some(r => r.parent_roles.includes(role.name));
  if (!hasChildren) {
    return '#8b5cf6'; // Purple for leaf
  }

  return '#6366f1'; // Indigo for intermediate
};

export function RoleGraph({ roles, onRoleClick }: RoleGraphProps) {
  const [selectedRoleName, setSelectedRoleName] = useState<string | null>(null);

  // Filter roles based on selection
  const filteredRoles = useMemo(() => {
    if (!selectedRoleName) {
      return roles;
    }

    const selectedRole = roles.find(r => r.name === selectedRoleName);
    if (!selectedRole) {
      return roles;
    }

    // Get parents
    const parents = selectedRole.parent_roles;

    // Get children (roles that have this role as a parent)
    const children = roles
      .filter(r => r.parent_roles.includes(selectedRoleName))
      .map(r => r.name);

    // Include selected role + parents + children
    const relevantRoles = new Set([selectedRoleName, ...parents, ...children]);

    return roles.filter(r => relevantRoles.has(r.name));
  }, [roles, selectedRoleName]);

  // Create nodes from filtered roles
  const initialNodes: Node[] = useMemo(() => {
    return filteredRoles.map((role) => {
      const isSelected = role.name === selectedRoleName;

      return {
        id: role.name,
        data: {
          label: (
            <div className="text-center">
              <div className="font-semibold text-sm">{role.name}</div>
              {role.comment && (
                <div className="text-xs text-gray-600 mt-1 truncate max-w-[180px]">
                  {role.comment}
                </div>
              )}
            </div>
          )
        },
        position: { x: 0, y: 0 },
        style: {
          background: isSelected ? '#fef3c7' : '#ffffff',
          border: `${isSelected ? '4' : '3'}px solid ${getNodeColor(role, roles)}`,
          borderRadius: '12px',
          padding: '12px 16px',
          width: nodeWidth,
          minHeight: nodeHeight,
          boxShadow: isSelected
            ? '0 10px 15px -3px rgb(0 0 0 / 0.2), 0 4px 6px -4px rgb(0 0 0 / 0.1)'
            : '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
          cursor: 'pointer',
          transition: 'all 0.2s ease',
        },
      };
    });
  }, [filteredRoles, selectedRoleName, roles]);

  // Create edges from parent relationships
  const initialEdges: Edge[] = useMemo(() => {
    const edges: Edge[] = [];
    filteredRoles.forEach((role) => {
      role.parent_roles.forEach((parent) => {
        // Only create edge if parent is in filtered roles
        if (filteredRoles.some(r => r.name === parent)) {
          edges.push({
            id: `${parent}-${role.name}`,
            source: parent,
            target: role.name,
            type: 'smoothstep',
            animated: role.name === selectedRoleName || parent === selectedRoleName,
            style: {
              stroke: role.name === selectedRoleName || parent === selectedRoleName
                ? '#3b82f6'
                : '#94a3b8',
              strokeWidth: role.name === selectedRoleName || parent === selectedRoleName
                ? 3
                : 2,
            },
            markerEnd: {
              type: MarkerType.ArrowClosed,
              width: 20,
              height: 20,
              color: role.name === selectedRoleName || parent === selectedRoleName
                ? '#3b82f6'
                : '#94a3b8',
            },
          });
        }
      });
    });
    return edges;
  }, [filteredRoles, selectedRoleName]);

  // Apply dagre layout
  const { nodes: layoutedNodes, edges: layoutedEdges } = useMemo(() => {
    return getLayoutedElements(initialNodes, initialEdges);
  }, [initialNodes, initialEdges]);

  const [nodes, setNodes, onNodesChange] = useNodesState(layoutedNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(layoutedEdges);

  // Update layout when filtered roles change
  useEffect(() => {
    const { nodes: newNodes, edges: newEdges } = getLayoutedElements(initialNodes, initialEdges);
    setNodes(newNodes);
    setEdges(newEdges);
  }, [initialNodes, initialEdges, setNodes, setEdges]);

  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      const role = roles.find((r) => r.name === node.id);
      if (role) {
        setSelectedRoleName(node.id);
        onRoleClick(role);
      }
    },
    [roles, onRoleClick]
  );

  const handleReset = useCallback(() => {
    setSelectedRoleName(null);
  }, []);

  if (roles.length === 0) {
    return (
      <div className="flex h-96 items-center justify-center rounded-lg border border-slate-200 bg-white">
        <p className="text-sm text-slate-500">No roles to display</p>
      </div>
    );
  }

  return (
    <div className="relative h-[700px] rounded-lg border border-slate-200 bg-white shadow-sm">
      {selectedRoleName && (
        <div className="absolute right-4 top-4 z-10">
          <Button
            variant="outline"
            size="sm"
            onClick={handleReset}
            className="shadow-md"
          >
            <RotateCcw className="mr-2 h-4 w-4" />
            Show All Roles
          </Button>
        </div>
      )}
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.5}
        maxZoom={1.5}
        defaultViewport={{ x: 0, y: 0, zoom: 0.8 }}
        attributionPosition="bottom-right"
      >
        <Background color="#f1f5f9" gap={16} />
        <Controls />
        <MiniMap
          nodeColor={(node) => {
            const role = roles.find(r => r.name === node.id);
            return role ? getNodeColor(role, roles) : '#94a3b8';
          }}
          nodeStrokeWidth={3}
          zoomable
          pannable
        />
      </ReactFlow>
    </div>
  );
}
