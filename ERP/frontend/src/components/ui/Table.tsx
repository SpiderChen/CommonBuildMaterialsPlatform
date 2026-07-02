import { Fragment, type ReactNode } from "react";
import { cx } from "./utils";

export type SimpleTableColumn<T> = {
  key: string;
  title: ReactNode;
  render: (row: T, index: number) => ReactNode;
  align?: "left" | "center" | "right";
  width?: string;
};

export type SimpleTableProps<T> = {
  columns: SimpleTableColumn<T>[];
  data: T[];
  rowKey: (row: T) => string | number;
  className?: string;
  emptyText?: ReactNode;
};

export function SimpleTable<T>({ columns, data, rowKey, className, emptyText = "暂无数据" }: SimpleTableProps<T>) {
  return (
    <table className={cx("ui-simple-table", className)} data-slot="ui-simple-table">
      <thead>
        <tr>
          {columns.map((column) => (
            <th key={column.key} style={{ textAlign: column.align || "left", width: column.width }}>
              {column.title}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {data.map((row, index) => (
          <tr key={rowKey(row)}>
            {columns.map((column) => (
              <td key={column.key} style={{ textAlign: column.align || "left" }}>
                {column.render(row, index)}
              </td>
            ))}
          </tr>
        ))}
        {!data.length ? (
          <tr>
            <td colSpan={columns.length}>{emptyText}</td>
          </tr>
        ) : null}
      </tbody>
    </table>
  );
}

export type KeyValueTableCell = {
  label: ReactNode;
  value: ReactNode;
};

export type KeyValueTableProps = {
  rows: KeyValueTableCell[][];
  className?: string;
};

export function KeyValueTable({ rows, className }: KeyValueTableProps) {
  return (
    <table className={cx("ui-key-value-table", className)} data-slot="ui-key-value-table">
      <tbody>
        {rows.map((row, rowIndex) => (
          <tr key={rowIndex}>
            {row.map((cell, cellIndex) => (
              <Fragment key={cellIndex}>
                <th>{cell.label}</th>
                <td>{cell.value}</td>
              </Fragment>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
