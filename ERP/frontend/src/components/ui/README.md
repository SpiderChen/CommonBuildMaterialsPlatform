# UI Component Library

This folder owns themeable primitives used by business views.

Theme variables live in `ui.css` under `:root`. Start theme changes there:

- `--ui-color-primary`
- `--ui-color-primary-contrast`
- `--ui-color-surface`
- `--ui-color-text`
- `--ui-color-muted`
- `--ui-color-border`
- `--ui-color-border-soft`
- `--ui-radius-sm`
- `--ui-radius-md`
- `--ui-shadow-dialog`
- `--ui-dialog-backdrop`
- `--ui-dialog-surface`
- `--ui-dialog-text`
- `--ui-dialog-border`
- `--ui-dialog-radius`
- `--ui-dialog-shadow`
- `--ui-dialog-max-width-sm`
- `--ui-dialog-max-width-md`
- `--ui-dialog-max-width-lg`
- `--ui-dialog-max-width-xl`
- `--ui-dialog-max-width-wide`
- `--ui-dialog-header-padding`
- `--ui-dialog-body-padding`
- `--ui-dialog-footer-padding`

Use `UiThemeProvider` for a scoped runtime theme:

```tsx
<UiThemeProvider
  theme={{
    colorPrimary: "#2458a6",
    dialogRadius: "10px",
    dialogShadow: "0 24px 60px rgba(24, 39, 68, .22)",
    dialogMaxWidthWide: "1040px"
  }}
>
  <App />
</UiThemeProvider>
```

Use `Dialog` for all modal surfaces, `DialogForm` for forms inside dialogs,
and `ActionDialog` when a row or toolbar button opens a dialog. Avoid writing
new `modal-backdrop`, `modal-panel`, `dialog-backdrop`, `dialog-panel`, or
`lab-form dialog-form` markup in feature modules.

Use `Button`, `ButtonLink`, `IconButton`, `ChipButton`, and `BareButton` for
actions. `BareButton` is for fully custom App Shell controls that should still
come through the component library. Use `Panel`, `Card`, `SelectableCard`,
`EmptyState`, `ActionGroup`, `FormGrid`, `LoginForm`, `QuickForm`,
`InlineForm`, `WorkflowForm`, `SystemForm`, and `FormActions` for page surfaces
and layouts instead of adding feature-specific shell classes.

Use `Field`, `IconField`, `TextInput`, `SelectInput`, and `TextAreaInput` for
form controls when touching forms. `IconField` is for search boxes, site
switchers, and other icon + control shells. These controls carry `data-slot`
markers so global resets do not override component-library theming.

Use `LayoutRegion`, `ViewStack`, `SectionGrid`, `SectionHeader`, `SplitRow`,
`MetricList`, and `ChipList` for repeated layout shells instead of writing raw
`main`, `aside`, `header`, `section`, `view-stack`, `grid-12`, `between`,
`metric-list`, or permission chip-list wrappers in feature modules.

Use `Message` through `useMessage()` for lightweight, non-blocking feedback
such as save, copy, refresh, and status-change results. Use `MessageBox`
through `useMessageBox()` for blocking errors and failures that require the
user's attention. Keep `FeedbackBanner` for rare inline, contextual guidance
that must stay attached to a specific panel.

Use `SimpleTable` for compact detail tables and `KeyValueTable` for report or
print-style key/value tables. Use `DataTable` for searchable/paginated resource
lists. `DataTable`, `HeroDateField`, `KpiCard`, and `StatusChip` are owned by
this library; root-level files only re-export them for compatibility.
`Panel` defaults to a `section`; pass `as="div"` when replacing an existing div
wrapper. New buttons, button groups, form grids, and form action rows should
come from this library.
