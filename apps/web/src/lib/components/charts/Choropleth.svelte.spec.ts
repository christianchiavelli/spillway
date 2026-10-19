import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';

import type { ResolvedPathname } from '$app/types';

import Choropleth from './Choropleth.svelte';

// Three squares in a row, in the map's own units.
const square = (x: number) => `M${x} 0L${x + 10} 0L${x + 10} 10L${x} 10Z`;
const regions = [
  { key: 'A', label: 'Most', value: 100, d: square(0), labelAt: [5, 5] as const },
  { key: 'B', label: 'Half', value: 50, d: square(10), labelAt: [15, 5] as const },
  { key: 'C', label: 'Little', value: 5, d: square(20), labelAt: [25, 5] as const },
].map((r) => ({ ...r, href: `/plants?subsystem=${r.key}` as ResolvedPathname }));

const fillOf = (container: HTMLElement, key: string) =>
  container.querySelector(`a[href$="=${key}"] path`)?.getAttribute('class') ?? '';

describe('Choropleth', () => {
  it('shades each region by its share of the largest, the most the darkest', async () => {
    const { container } = await render(Choropleth, { width: 30, height: 10, regions });

    expect(fillOf(container, 'A')).toContain('fill-chart-scale-4');
    expect(fillOf(container, 'B')).toContain('fill-chart-scale-3');
    expect(fillOf(container, 'C')).toContain('fill-chart-scale-1');
    expect(container.querySelector('svg')?.textContent).toContain('Little');
  });

  it('is left to the figures beside it for a screen reader and the keyboard', async () => {
    const { container } = await render(Choropleth, { width: 30, height: 10, regions });

    expect(container.querySelector('[aria-hidden="true"] svg')).not.toBeNull();
    for (const link of container.querySelectorAll('a')) {
      expect(link.getAttribute('tabindex')).toBe('-1');
    }
  });

  it('outlines the region pointed at and fades the others, and says which', async () => {
    const onactive = vi.fn();
    const { container, rerender } = await render(Choropleth, {
      width: 30,
      height: 10,
      regions,
      onactive,
    });
    expect(container.querySelectorAll('.opacity-40')).toHaveLength(0);

    await rerender({ active: 'B' });
    expect(fillOf(container, 'B')).not.toContain('opacity-40');
    expect(container.querySelectorAll('.opacity-40')).toHaveLength(2);
    expect(container.querySelectorAll('.stroke-text-primary-default')).toHaveLength(1);

    container
      .querySelector('a[href$="=C"]')
      ?.dispatchEvent(new PointerEvent('pointerenter', { bubbles: false }));
    expect(onactive).toHaveBeenLastCalledWith('C');
  });
});
