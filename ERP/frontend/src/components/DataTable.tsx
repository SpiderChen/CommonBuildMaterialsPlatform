import { ChevronLeft, ChevronRight } from "lucide-react";
import { type ReactNode, useEffect, useMemo, useState } from "react";

export type DataTableColumn<T> = {
  key: string;
  title: ReactNode;
  render: (row: T, index: number) => ReactNode;
  align?: "left" | "center" | "right";
  width?: string;
};

type DataTableProps<T> = {
  title?: ReactNode;
  data: T[];
  columns: DataTableColumn<T>[];
  rowKey: (row: T) => string | number;
  emptyText?: string;
  headerAction?: ReactNode;
  pageSize?: number;
  pageSizeOptions?: number[];
  showPagination?: boolean;
};

export function DataTable<T>({
  title,
  data,
  columns,
  rowKey,
  emptyText = "暂无数据",
  headerAction,
  pageSize = 10,
  pageSizeOptions = [10, 20, 50],
  showPagination = true
}: DataTableProps<T>) {
  const [currentPage, setCurrentPage] = useState(1);
  const [currentPageSize, setCurrentPageSize] = useState(pageSize);
  const total = data.length;
  const totalPages = Math.max(1, Math.ceil(total / currentPageSize));
  const page = Math.min(currentPage, totalPages);
  const paginationEnabled = showPagination && total > 0;
  const start = paginationEnabled ? (page - 1) * currentPageSize : 0;
  const end = paginationEnabled ? Math.min(start + currentPageSize, total) : total;
  const visibleRows = useMemo(() => data.slice(start, end), [data, start, end]);

  useEffect(() => {
    setCurrentPageSize(pageSize);
  }, [pageSize]);

  useEffect(() => {
    if (currentPage > totalPages) {
      setCurrentPage(totalPages);
    }
  }, [currentPage, totalPages]);

  function changePageSize(value: string) {
    setCurrentPageSize(Number(value));
    setCurrentPage(1);
  }

  return (
    <div className="data-table">
      {title || headerAction ? (
        <div className="data-table-head">
          {title ? <h3>{title}</h3> : <span />}
          {headerAction}
        </div>
      ) : null}
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              {columns.map((column) => (
                <th key={column.key} style={{ width: column.width, textAlign: column.align || "left" }}>
                  {column.title}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {visibleRows.map((row, index) => (
              <tr key={rowKey(row)}>
                {columns.map((column) => (
                  <td key={column.key} style={{ textAlign: column.align || "left" }}>
                    {column.render(row, start + index)}
                  </td>
                ))}
              </tr>
            ))}
            {!total ? (
              <tr>
                <td className="data-table-empty" colSpan={columns.length}>{emptyText}</td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
      {paginationEnabled ? (
        <div className="data-table-pagination">
          <span className="pagination-info">
            {start + 1}-{end} / {total}
          </span>
          <label className="pagination-size">
            <span>每页</span>
            <select value={currentPageSize} onChange={(event) => changePageSize(event.target.value)}>
              {pageSizeOptions.map((item) => <option key={item} value={item}>{item}</option>)}
            </select>
          </label>
          <div className="pagination-steps">
            <button className="icon-button" type="button" aria-label="上一页" disabled={page <= 1} onClick={() => setCurrentPage((value) => Math.max(1, value - 1))}>
              <ChevronLeft size={16} />
            </button>
            <span>{page} / {totalPages}</span>
            <button className="icon-button" type="button" aria-label="下一页" disabled={page >= totalPages} onClick={() => setCurrentPage((value) => Math.min(totalPages, value + 1))}>
              <ChevronRight size={16} />
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
