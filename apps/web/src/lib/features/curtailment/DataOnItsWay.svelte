<script lang="ts">
  import EmptyState from '#lib/components/ui/EmptyState.svelte';
  import TurbinesArt from '#lib/components/art/TurbinesArt.svelte';
  import { m } from '#lib/paraglide/messages.js';
</script>

<!-- What a page shows in the minutes a fresh start takes to load the ONS's files. -->
<EmptyState headingLevel={2} title={m.waiting_title()} description={m.waiting_body()}>
  {#snippet art()}
    <TurbinesArt />
  {/snippet}
</EmptyState>
