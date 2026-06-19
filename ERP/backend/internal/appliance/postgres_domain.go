package appliance

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
)

const postgresDomainSingletonRowID = "__singleton"

type postgresDomainRow struct {
	Resource string
	RowID    string
	Ordinal  int
	Payload  json.RawMessage
	Checksum string
}

func postgresDomainSchemaSQL() string {
	return `
		create table if not exists cbmp_domain_rows (
			resource text not null,
			row_id text not null,
			ordinal integer not null,
			payload jsonb not null,
			snapshot_checksum text not null,
			updated_at timestamptz not null default now(),
			primary key (resource, row_id)
		);
		create index if not exists idx_cbmp_domain_rows_resource_ordinal on cbmp_domain_rows (resource, ordinal);

		create table if not exists cbmp_domain_status (
			id text primary key,
			resource_count integer not null,
			row_count integer not null,
			snapshot_checksum text not null,
			refreshed_at timestamptz not null default now()
		);
	`
}

func (s *PostgresStore) refreshDomainRows(ctx context.Context, tx pgx.Tx, data AppData, snapshotChecksum string) error {
	rows, err := domainRowsFromAppData(data, snapshotChecksum)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "delete from cbmp_domain_rows"); err != nil {
		return err
	}
	var batch pgx.Batch
	for _, row := range rows {
		batch.Queue(`
			insert into cbmp_domain_rows (resource, row_id, ordinal, payload, snapshot_checksum, updated_at)
			values ($1, $2, $3, $4, $5, now())
		`, row.Resource, row.RowID, row.Ordinal, string(row.Payload), row.Checksum)
	}
	if batch.Len() > 0 {
		results := tx.SendBatch(ctx, &batch)
		for i := 0; i < batch.Len(); i++ {
			if _, err := results.Exec(); err != nil {
				_ = results.Close()
				return err
			}
		}
		if err := results.Close(); err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `
		insert into cbmp_domain_status (id, resource_count, row_count, snapshot_checksum, refreshed_at)
		values ('default', $1, $2, $3, now())
		on conflict (id)
		do update set resource_count = excluded.resource_count, row_count = excluded.row_count, snapshot_checksum = excluded.snapshot_checksum, refreshed_at = now()
	`, len(appDataDomainResources()), len(rows), snapshotChecksum)
	return err
}

func (s *PostgresStore) loadDomainRows(ctx context.Context) (AppData, bool, error) {
	var rowCount int
	err := s.pool.QueryRow(ctx, `select row_count from cbmp_domain_status where id = 'default'`).Scan(&rowCount)
	if err == pgx.ErrNoRows {
		return AppData{}, false, nil
	}
	if err != nil {
		return AppData{}, false, err
	}
	if rowCount == 0 {
		return AppData{}, false, nil
	}
	rows, err := s.pool.Query(ctx, `
		select resource, row_id, ordinal, payload, snapshot_checksum
		from cbmp_domain_rows
		order by resource, ordinal, row_id
	`)
	if err != nil {
		return AppData{}, false, err
	}
	defer rows.Close()
	var domainRows []postgresDomainRow
	for rows.Next() {
		var row postgresDomainRow
		if err := rows.Scan(&row.Resource, &row.RowID, &row.Ordinal, &row.Payload, &row.Checksum); err != nil {
			return AppData{}, false, err
		}
		domainRows = append(domainRows, row)
	}
	if err := rows.Err(); err != nil {
		return AppData{}, false, err
	}
	if len(domainRows) == 0 {
		return AppData{}, false, nil
	}
	data, err := appDataFromDomainRows(domainRows)
	return data, true, err
}

func domainRowsFromAppData(data AppData, snapshotChecksum string) ([]postgresDomainRow, error) {
	value := reflect.ValueOf(data)
	typ := value.Type()
	rows := make([]postgresDomainRow, 0)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		resource := appDataResourceName(field)
		if resource == "" {
			continue
		}
		fieldValue := value.Field(i)
		if fieldValue.Kind() == reflect.Slice {
			for j := 0; j < fieldValue.Len(); j++ {
				item := fieldValue.Index(j)
				payload, err := json.Marshal(item.Interface())
				if err != nil {
					return nil, fmt.Errorf("marshal domain row %s[%d]: %w", resource, j, err)
				}
				rows = append(rows, postgresDomainRow{
					Resource: resource,
					RowID:    domainRowID(item, j),
					Ordinal:  j,
					Payload:  payload,
					Checksum: snapshotChecksum,
				})
			}
			continue
		}
		payload, err := json.Marshal(fieldValue.Interface())
		if err != nil {
			return nil, fmt.Errorf("marshal domain singleton %s: %w", resource, err)
		}
		rows = append(rows, postgresDomainRow{
			Resource: resource,
			RowID:    postgresDomainSingletonRowID,
			Ordinal:  0,
			Payload:  payload,
			Checksum: snapshotChecksum,
		})
	}
	return rows, nil
}

func appDataFromDomainRows(rows []postgresDomainRow) (AppData, error) {
	var data AppData
	value := reflect.ValueOf(&data).Elem()
	fields := appDataDomainFieldIndex(value.Type())
	for _, row := range rows {
		index, ok := fields[row.Resource]
		if !ok {
			continue
		}
		field := value.Field(index)
		if field.Kind() == reflect.Slice {
			item := reflect.New(field.Type().Elem())
			if err := json.Unmarshal(row.Payload, item.Interface()); err != nil {
				return AppData{}, fmt.Errorf("unmarshal domain row %s/%s: %w", row.Resource, row.RowID, err)
			}
			field.Set(reflect.Append(field, item.Elem()))
			continue
		}
		target := reflect.New(field.Type())
		if err := json.Unmarshal(row.Payload, target.Interface()); err != nil {
			return AppData{}, fmt.Errorf("unmarshal domain singleton %s: %w", row.Resource, err)
		}
		field.Set(target.Elem())
	}
	if data.Next == nil {
		data.Next = map[string]int64{}
	}
	return data, nil
}

func domainRowCount(data AppData) int {
	rows, err := domainRowsFromAppData(data, "")
	if err != nil {
		return 0
	}
	return len(rows)
}

func appDataDomainResources() []string {
	value := reflect.TypeOf(AppData{})
	out := make([]string, 0, value.NumField())
	for i := 0; i < value.NumField(); i++ {
		if resource := appDataResourceName(value.Field(i)); resource != "" {
			out = append(out, resource)
		}
	}
	return out
}

func appDataDomainFieldIndex(typ reflect.Type) map[string]int {
	out := map[string]int{}
	for i := 0; i < typ.NumField(); i++ {
		if resource := appDataResourceName(typ.Field(i)); resource != "" {
			out[resource] = i
		}
	}
	return out
}

func appDataResourceName(field reflect.StructField) string {
	if field.PkgPath != "" {
		return ""
	}
	tag := field.Tag.Get("json")
	if tag == "-" {
		return ""
	}
	if tag != "" {
		name := strings.Split(tag, ",")[0]
		if name != "" {
			return name
		}
	}
	return field.Name
}

func domainRowID(value reflect.Value, ordinal int) string {
	identity := domainRowIdentity(value)
	if identity == "" {
		return fmt.Sprintf("%06d", ordinal)
	}
	return fmt.Sprintf("%06d:%s", ordinal, identity)
}

func domainRowIdentity(value reflect.Value) string {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return ""
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return ""
	}
	if field := value.FieldByName("ID"); field.IsValid() && field.CanInterface() {
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() != 0 {
				return fmt.Sprintf("%d", field.Int())
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if field.Uint() != 0 {
				return fmt.Sprintf("%d", field.Uint())
			}
		}
	}
	for _, name := range []string{"Code", "InvoiceNo", "SubmissionNo", "OrderNo", "DispatchNo", "TicketNo", "FrameNo", "DeviceNo", "PlateNo", "Username", "LicenseID", "RunNo", "EventNo"} {
		field := value.FieldByName(name)
		if field.IsValid() && field.CanInterface() && field.Kind() == reflect.String {
			if item := strings.TrimSpace(field.String()); item != "" {
				return item
			}
		}
	}
	return ""
}
