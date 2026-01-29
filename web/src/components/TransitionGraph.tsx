import { useEffect, useRef, useCallback, useState } from 'react';
import { motion } from 'framer-motion';
import * as d3 from 'd3';
import type { SetEdge, Track } from '../types';

interface GraphNode extends d3.SimulationNodeDatum {
  id: string;
  title: string;
  bpm: number;
  key: string;
  energy: number;
}

interface GraphLink extends d3.SimulationLinkDatum<GraphNode> {
  score: number;
  reason: string;
}

interface TransitionGraphProps {
  tracks: Record<string, Track>;
  edges: SetEdge[];
  width?: number;
  height?: number;
  selectedTrackId?: string;
  onSelectTrack?: (id: string) => void;
}

export function TransitionGraph({
  tracks,
  edges,
  width = 400,
  height = 300,
  selectedTrackId,
  onSelectTrack,
}: TransitionGraphProps) {
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width, height });
  const [hoveredNode, setHoveredNode] = useState<GraphNode | null>(null);

  // Resize observer
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width: w, height: h } = entry.contentRect;
        if (w > 0 && h > 0) {
          setDimensions({ width: w, height: h });
        }
      }
    });

    resizeObserver.observe(container);
    return () => resizeObserver.disconnect();
  }, []);

  const getNodeColor = useCallback((energy: number): string => {
    if (energy >= 8) return '#ef4444';
    if (energy >= 6) return '#f59e0b';
    if (energy >= 4) return '#3b82f6';
    return '#22c55e';
  }, []);

  const getKeyColor = useCallback((key: string): string => {
    // Map Camelot keys to hue values for visual grouping
    const keyNum = parseInt(key.replace(/[AB]/g, '')) || 1;
    const hue = (keyNum - 1) * 30;
    return `hsl(${hue}, 70%, 60%)`;
  }, []);

  // D3 Force Simulation
  useEffect(() => {
    const svg = d3.select(svgRef.current);
    if (!svg.node()) return;

    const { width: w, height: h } = dimensions;

    // Clear previous
    svg.selectAll('*').remove();

    // Prepare data
    const trackList = Object.values(tracks);
    const nodes: GraphNode[] = trackList.map((t) => ({
      id: t.id,
      title: t.title,
      bpm: t.bpm,
      key: t.key,
      energy: t.energy,
    }));

    const links: GraphLink[] = edges
      .filter((e) => tracks[e.from] && tracks[e.to])
      .map((e) => ({
        source: e.from,
        target: e.to,
        score: e.score,
        reason: e.reason,
      }));

    if (nodes.length === 0) return;

    // Create container group for zoom
    const g = svg.append('g');

    // Add zoom behavior
    const zoom = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.3, 3])
      .on('zoom', (event) => {
        g.attr('transform', event.transform);
      });

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    svg.call(zoom as any);

    // Arrow marker for directed edges
    svg
      .append('defs')
      .append('marker')
      .attr('id', 'arrowhead')
      .attr('viewBox', '-0 -5 10 10')
      .attr('refX', 20)
      .attr('refY', 0)
      .attr('orient', 'auto')
      .attr('markerWidth', 6)
      .attr('markerHeight', 6)
      .append('path')
      .attr('d', 'M 0,-5 L 10,0 L 0,5')
      .attr('fill', 'var(--color-text-muted)');

    // Glow filter
    const defs = svg.select('defs');
    const filter = defs.append('filter').attr('id', 'glow');
    filter.append('feGaussianBlur').attr('stdDeviation', '3').attr('result', 'coloredBlur');
    const feMerge = filter.append('feMerge');
    feMerge.append('feMergeNode').attr('in', 'coloredBlur');
    feMerge.append('feMergeNode').attr('in', 'SourceGraphic');

    // Force simulation
    const simulation = d3
      .forceSimulation<GraphNode>(nodes)
      .force(
        'link',
        d3
          .forceLink<GraphNode, GraphLink>(links)
          .id((d) => d.id)
          .distance(80)
          .strength((d) => d.score / 15)
      )
      .force('charge', d3.forceManyBody().strength(-200))
      .force('center', d3.forceCenter(w / 2, h / 2))
      .force('collision', d3.forceCollide().radius(30));

    // Draw links
    const link = g
      .append('g')
      .attr('class', 'links')
      .selectAll('line')
      .data(links)
      .join('line')
      .attr('stroke', (d) => {
        const score = d.score;
        if (score >= 8) return 'rgba(34, 197, 94, 0.6)';
        if (score >= 6) return 'rgba(59, 130, 246, 0.6)';
        return 'rgba(156, 163, 175, 0.4)';
      })
      .attr('stroke-width', (d) => Math.max(1, d.score / 3))
      .attr('marker-end', 'url(#arrowhead)')
      .style('cursor', 'pointer');

    // Link labels (scores)
    const linkLabels = g
      .append('g')
      .attr('class', 'link-labels')
      .selectAll('text')
      .data(links)
      .join('text')
      .attr('font-size', '8px')
      .attr('fill', 'var(--color-text-muted)')
      .attr('text-anchor', 'middle')
      .text((d) => d.score.toFixed(1));

    // Draw nodes
    const node = g
      .append('g')
      .attr('class', 'nodes')
      .selectAll('g')
      .data(nodes)
      .join('g')
      .style('cursor', 'pointer')
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .call(
        d3
          .drag<SVGGElement, GraphNode>()
          .on('start', (event, d) => {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
          })
          .on('drag', (event, d) => {
            d.fx = event.x;
            d.fy = event.y;
          })
          .on('end', (event, d) => {
            if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
          }) as any
      );

    // Node circles
    node
      .append('circle')
      .attr('r', (d) => 10 + d.energy)
      .attr('fill', (d) => getNodeColor(d.energy))
      .attr('stroke', (d) => (d.id === selectedTrackId ? '#ffffff' : getKeyColor(d.key)))
      .attr('stroke-width', (d) => (d.id === selectedTrackId ? 3 : 2))
      .attr('filter', (d) => (d.id === selectedTrackId ? 'url(#glow)' : null))
      .on('click', (_, d) => {
        onSelectTrack?.(d.id);
      })
      .on('mouseenter', (_, d) => {
        setHoveredNode(d);
      })
      .on('mouseleave', () => {
        setHoveredNode(null);
      });

    // Node labels
    node
      .append('text')
      .attr('dy', (d) => 18 + d.energy)
      .attr('text-anchor', 'middle')
      .attr('font-size', '9px')
      .attr('fill', 'var(--color-text-secondary)')
      .text((d) => d.title.slice(0, 12));

    // BPM/Key indicator
    node
      .append('text')
      .attr('dy', 3)
      .attr('text-anchor', 'middle')
      .attr('font-size', '7px')
      .attr('font-weight', 'bold')
      .attr('fill', '#ffffff')
      .text((d) => d.key);

    // Simulation tick
    simulation.on('tick', () => {
      link
        .attr('x1', (d) => (d.source as GraphNode).x ?? 0)
        .attr('y1', (d) => (d.source as GraphNode).y ?? 0)
        .attr('x2', (d) => (d.target as GraphNode).x ?? 0)
        .attr('y2', (d) => (d.target as GraphNode).y ?? 0);

      linkLabels
        .attr('x', (d) => ((d.source as GraphNode).x! + (d.target as GraphNode).x!) / 2)
        .attr('y', (d) => ((d.source as GraphNode).y! + (d.target as GraphNode).y!) / 2);

      node.attr('transform', (d) => `translate(${d.x ?? 0},${d.y ?? 0})`);
    });

    // Cleanup
    return () => {
      simulation.stop();
    };
  }, [tracks, edges, dimensions, selectedTrackId, onSelectTrack, getNodeColor, getKeyColor]);

  return (
    <motion.div
      ref={containerRef}
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
      style={{
        width: '100%',
        height,
        position: 'relative',
        background: 'var(--color-bg-tertiary)',
        borderRadius: 8,
        overflow: 'hidden',
      }}
    >
      <svg ref={svgRef} width={dimensions.width} height={dimensions.height} />

      {/* Legend */}
      <div
        style={{
          position: 'absolute',
          top: 8,
          left: 8,
          fontSize: '0.6rem',
          color: 'var(--color-text-muted)',
          background: 'rgba(0,0,0,0.5)',
          padding: '4px 8px',
          borderRadius: 4,
        }}
      >
        <div style={{ marginBottom: 4 }}>
          <strong>Node size</strong> = Energy
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <span>
            <span style={{ color: '#22c55e' }}>●</span> Low
          </span>
          <span>
            <span style={{ color: '#3b82f6' }}>●</span> Mid
          </span>
          <span>
            <span style={{ color: '#f59e0b' }}>●</span> High
          </span>
          <span>
            <span style={{ color: '#ef4444' }}>●</span> Peak
          </span>
        </div>
      </div>

      {/* Hover tooltip */}
      {hoveredNode && (
        <motion.div
          initial={{ opacity: 0, y: 5 }}
          animate={{ opacity: 1, y: 0 }}
          style={{
            position: 'absolute',
            bottom: 8,
            left: 8,
            right: 8,
            background: 'var(--color-bg-secondary)',
            border: '1px solid var(--color-border)',
            borderRadius: 6,
            padding: '8px 12px',
            fontSize: '0.75rem',
          }}
        >
          <div style={{ fontWeight: 600 }}>{hoveredNode.title}</div>
          <div style={{ color: 'var(--color-text-secondary)', marginTop: 2 }}>
            {hoveredNode.bpm} BPM · {hoveredNode.key} · Energy {hoveredNode.energy}
          </div>
        </motion.div>
      )}

      {/* Zoom instructions */}
      <div
        style={{
          position: 'absolute',
          bottom: 8,
          right: 8,
          fontSize: '0.55rem',
          color: 'var(--color-text-muted)',
          opacity: 0.6,
        }}
      >
        Scroll to zoom · Drag to pan
      </div>
    </motion.div>
  );
}
