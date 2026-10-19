<script lang="ts">
  type Props = {
    value: number;
    /** The value of a full bar. */
    max: number;
    /** The class that paints it. */
    fill: string;
    width?: number;
  };

  let { value, max, fill, width = 96 }: Props = $props();

  const w = $derived(Math.max(2, Math.round((value / Math.max(max, Number.EPSILON)) * width)));
  const r = $derived(Math.min(3, w / 2));
</script>

<!-- A bar in a table cell: its length is the row's share of the largest. -->
<svg {width} height="8" aria-hidden="true" class="shrink-0">
  <path d="M0,0H{w - r}A{r},{r} 0 0 1 {w},{r}V{8 - r}A{r},{r} 0 0 1 {w - r},8H0Z" class={fill} />
</svg>
