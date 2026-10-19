<script lang="ts">
  import SegmentedControl from '#lib/components/ui/SegmentedControl.svelte';
  import SelectField from '#lib/components/ui/SelectField.svelte';
  import TextField from '#lib/components/ui/TextField.svelte';
  import { sources, subsystems } from '#lib/features/curtailment/grid.ts';
  import type { FieldErrors } from '#lib/forms/errors.ts';

  import { type AlertField, alertError, limits } from './alerts.ts';

  type Props = {
    /** What was typed, sent back after a failed submit so nothing is lost. */
    values?: Partial<Record<AlertField, string>> | undefined;
    errors?: FieldErrors<AlertField> | undefined;
  };

  let { values = {}, errors = {} }: Props = $props();
</script>

<!--
  The fields only: the page owns the form, its action and what happens on
  submit. Constraints come from the contract and are there for assistive
  technology and the keyboard; the form posts with novalidate, so the server
  checks and words every rule, with JavaScript or without.
-->
<!--
  One field under another, the source first: it is the first thing to choose,
  and the form shares its row with the preview of what it would catch.
-->
<div class="flex flex-col gap-600">
  <SegmentedControl
    id="alert-source"
    legend="Source"
    name="source"
    options={sources}
    value={values.source ?? 'wind'}
    required
    error={alertError('source', errors.source)}
  />
  <TextField
    id="alert-name"
    label="Name"
    name="name"
    value={values.name}
    required
    maxlength={limits.nameMaxLength}
    autocomplete="off"
    error={alertError('name', errors.name)}
    hint="Unique among your alerts"
  />
  <SelectField
    id="alert-subsystem"
    label="Subsystem"
    name="subsystem"
    options={subsystems}
    value={values.subsystem}
    placeholder="Choose a subsystem"
    required
    error={alertError('subsystem', errors.subsystem)}
  />
  <TextField
    id="alert-threshold"
    label="Daily threshold"
    name="thresholdMwh"
    type="text"
    inputmode="numeric"
    value={values.thresholdMwh}
    suffix="MWh"
    required
    error={alertError('thresholdMwh', errors.thresholdMwh)}
    hint="Curtailed in one day"
  />
</div>
