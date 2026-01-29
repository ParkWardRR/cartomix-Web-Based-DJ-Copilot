import { useEffect, useRef, useState, useCallback } from 'react';
import { motion } from 'framer-motion';
import * as d3 from 'd3';
import type { Track } from '../types';

interface BPMKeyChartProps {
  tracks: Track[];
  height?: number;
  mode?: 'bpm' | 'key' | 'camelot';
}

// Camelot wheel order for proper key grouping
const CAMELOT_ORDER = ['1A', '1B', '2A', '2B', '3A', '3B', '4A', '4B', '5A', '5B', '6A', '6B', '7A', '7B', '8A', '8B', '9A', '9B', '10A', '10B', '11A', '11B', '12A', '12B'];

export function BPMKeyChart({ tracks, height = 160, mode = 'bpm' }: BPMKeyChartProps) {
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width: 300, height });
  const [hoveredBar, setHoveredBar] = useState<{ label: string; count: number; x: number; y: number } | null>(null);

  // Resize observer
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width } = entry.contentRect;
        if (width > 0) {
          setDimensions({ width, height });
        }
      }
    });

    resizeObserver.observe(container);
    return () => resizeObserver.disconnect();
  }, [height]);

  const getBarColor = useCallback((index: number, total: number): string => {
    if (mode === 'key' || mode === 'camelot') {
      // Use Camelot wheel colors
      const hue = (index / total) * 360;
      return `hsl(${hue}, 65%, 55%)`;
    }
    // BPM mode - gradient from blue to red
    const ratio = index / total;
    if (ratio < 0.33) return 'var(--color-success)';
    if (ratio < 0.66) return 'var(--color-primary)';
    return 'var(--color-error)';
  }, [mode]);

  useEffect(() => {
    const svg = d3.select(svgRef.current);
    if (!svg.node() || tracks.length === 0) return;

    const { width: w, height: h } = dimensions;
    const margin = { top: 20, right: 20, bottom: 40, left: 40 };
    const innerWidth = w - margin.left - margin.right;
    const innerHeight = h - margin.top - margin.bottom;

    svg.selectAll('*').remove();

    // Prepare data based on mode
    let data: { label: string; count: number }[] = [];

    if (mode === 'bpm') {
      // Group BPM into ranges
      const bpmRanges = [
        { min: 0, max: 100, label: '<100' },
        { min: 100, max: 110, label: '100-110' },
        { min: 110, max: 120, label: '110-120' },
        { min: 120, max: 128, label: '120-128' },
        { min: 128, max: 135, label: '128-135' },
        { min: 135, max: 145, label: '135-145' },
        { min: 145, max: 160, label: '145-160' },
        { min: 160, max: 999, label: '160+' },
      ];

      data = bpmRanges.map((range) => ({
        label: range.label,
        count: tracks.filter((t) => t.bpm >= range.min && t.bpm < range.max).length,
      }));
    } else {
      // Key distribution
      const keyCounts = new Map<string, number>();
      tracks.forEach((t) => {
        const key = t.key;
        keyCounts.set(key, (keyCounts.get(key) || 0) + 1);
      });

      // Sort by Camelot order
      const sortedKeys = Array.from(keyCounts.keys()).sort((a, b) => {
        const aIdx = CAMELOT_ORDER.indexOf(a);
        const bIdx = CAMELOT_ORDER.indexOf(b);
        return (aIdx === -1 ? 999 : aIdx) - (bIdx === -1 ? 999 : bIdx);
      });

      data = sortedKeys.map((key) => ({
        label: key,
        count: keyCounts.get(key) || 0,
      }));
    }

    // Filter out empty bars for cleaner visualization
    data = data.filter((d) => d.count > 0 || mode === 'bpm');

    if (data.length === 0) return;

    const g = svg.append('g').attr('transform', `translate(${margin.left},${margin.top})`);

    // Scales
    const x = d3
      .scaleBand()
      .domain(data.map((d) => d.label))
      .range([0, innerWidth])
      .padding(0.2);

    const y = d3
      .scaleLinear()
      .domain([0, d3.max(data, (d) => d.count) || 1])
      .nice()
      .range([innerHeight, 0]);

    // Grid lines
    g.append('g')
      .attr('class', 'grid')
      .selectAll('line')
      .data(y.ticks(4))
      .join('line')
      .attr('x1', 0)
      .attr('x2', innerWidth)
      .attr('y1', (d) => y(d))
      .attr('y2', (d) => y(d))
      .attr('stroke', 'var(--color-border)')
      .attr('stroke-dasharray', '2,2')
      .attr('opacity', 0.5);

    // Bars with animation
    g.selectAll('.bar')
      .data(data)
      .join('rect')
      .attr('class', 'bar')
      .attr('x', (d) => x(d.label) || 0)
      .attr('width', x.bandwidth())
      .attr('y', innerHeight)
      .attr('height', 0)
      .attr('rx', 3)
      .attr('fill', (_d, i) => getBarColor(i, data.length))
      .style('cursor', 'pointer')
      .on('mouseenter', function (event, d) {
        const rect = (event.target as SVGRectElement).getBoundingClientRect();
        const containerRect = containerRef.current?.getBoundingClientRect();
        if (containerRect) {
          setHoveredBar({
            label: d.label,
            count: d.count,
            x: rect.left - containerRect.left + rect.width / 2,
            y: rect.top - containerRect.top - 10,
          });
        }
        d3.select(this).attr('opacity', 0.8);
      })
      .on('mouseleave', function () {
        setHoveredBar(null);
        d3.select(this).attr('opacity', 1);
      })
      .transition()
      .duration(600)
      .delay((_, i) => i * 50)
      .ease(d3.easeElasticOut.amplitude(1).period(0.4))
      .attr('y', (d) => y(d.count))
      .attr('height', (d) => innerHeight - y(d.count));

    // X axis
    g.append('g')
      .attr('transform', `translate(0,${innerHeight})`)
      .call(d3.axisBottom(x).tickSize(0))
      .selectAll('text')
      .attr('font-size', '0.6rem')
      .attr('fill', 'var(--color-text-muted)')
      .attr('transform', 'rotate(-35)')
      .attr('text-anchor', 'end')
      .attr('dx', '-0.5em')
      .attr('dy', '0.5em');

    g.select('.domain').remove();

    // Y axis
    g.append('g')
      .call(d3.axisLeft(y).ticks(4).tickSize(-innerWidth))
      .selectAll('text')
      .attr('font-size', '0.6rem')
      .attr('fill', 'var(--color-text-muted)');

    g.selectAll('.tick line').attr('stroke', 'transparent');
    g.select('.domain').remove();

    // Title
    svg
      .append('text')
      .attr('x', w / 2)
      .attr('y', 14)
      .attr('text-anchor', 'middle')
      .attr('font-size', '0.7rem')
      .attr('font-weight', '600')
      .attr('fill', 'var(--color-text-secondary)')
      .text(mode === 'bpm' ? 'BPM Distribution' : 'Key Distribution');
  }, [tracks, dimensions, mode, getBarColor]);

  return (
    <motion.div
      ref={containerRef}
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
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

      {/* Hover tooltip */}
      {hoveredBar && (
        <motion.div
          initial={{ opacity: 0, y: 5 }}
          animate={{ opacity: 1, y: 0 }}
          style={{
            position: 'absolute',
            left: hoveredBar.x,
            top: hoveredBar.y,
            transform: 'translateX(-50%)',
            background: 'var(--color-bg-secondary)',
            border: '1px solid var(--color-border)',
            borderRadius: 4,
            padding: '4px 8px',
            fontSize: '0.7rem',
            pointerEvents: 'none',
            whiteSpace: 'nowrap',
            zIndex: 10,
          }}
        >
          <strong>{hoveredBar.label}</strong>: {hoveredBar.count} tracks
        </motion.div>
      )}
    </motion.div>
  );
}
