<script lang="ts">
  import EmptyState from '#lib/components/ui/EmptyState.svelte';
  import TurbinesArt from '#lib/components/art/TurbinesArt.svelte';
</script>

<!-- What a page shows in the minutes a fresh start takes to load the ONS's files. -->
<EmptyState
  headingLevel={2}
  title="The ONS's data is on its way"
  description="Spillway is reading the files the grid operator publishes, the latest month first. This takes a few minutes the first time. Reload the page in a moment."
>
  {#snippet art()}
    <TurbinesArt />
  {/snippet}
</EmptyState>
