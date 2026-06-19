package appliance

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

var postgresProjectionTableNames = []string{
	"cbmp_biz_customer_contacts",
	"cbmp_biz_customer_blacklists",
	"cbmp_biz_customer_profiles",
	"cbmp_biz_customer_complaints",
	"cbmp_biz_contract_attachments",
	"cbmp_biz_price_policies",
	"cbmp_biz_tax_rates",
	"cbmp_biz_orders",
	"cbmp_biz_order_lines",
	"cbmp_biz_dispatch_orders",
	"cbmp_biz_dispatch_schedules",
	"cbmp_biz_transport_settlements",
	"cbmp_biz_transport_settlement_items",
	"cbmp_biz_mix_designs",
	"cbmp_biz_mix_design_trial_runs",
	"cbmp_biz_laboratory_samples",
	"cbmp_biz_laboratory_tests",
	"cbmp_biz_laboratory_equipment",
	"cbmp_biz_laboratory_calibrations",
	"cbmp_biz_quality_exceptions",
	"cbmp_biz_production_batches",
	"cbmp_biz_scale_tickets",
	"cbmp_biz_scale_device_events",
	"cbmp_biz_delivery_sign_links",
	"cbmp_biz_delivery_signs",
	"cbmp_biz_delivery_sign_attachments",
	"cbmp_biz_sales_invoices",
	"cbmp_biz_red_letter_infos",
	"cbmp_biz_tax_gateway_submissions",
	"cbmp_biz_receivables",
	"cbmp_biz_payment_plans",
	"cbmp_biz_collection_tasks",
	"cbmp_biz_collection_templates",
	"cbmp_biz_collection_dispatches",
	"cbmp_biz_inventory_items",
	"cbmp_biz_inventory_flows",
	"cbmp_biz_vehicle_locations",
	"cbmp_biz_latest_locations",
	"cbmp_biz_device_protocol_frames",
}

type businessProjectionInsert struct {
	table   string
	columns []string
	values  []interface{}
}

func postgresProjectionSchemaSQL() string {
	return `
		create table if not exists cbmp_biz_customer_contacts (
			id bigint primary key,
			customer_id bigint not null,
			name text not null,
			phone text not null,
			role text not null,
			is_default boolean not null,
			status text not null
		);
		create index if not exists idx_cbmp_biz_customer_contacts_customer on cbmp_biz_customer_contacts (customer_id, status);

		create table if not exists cbmp_biz_customer_blacklists (
			id bigint primary key,
			customer_id bigint not null,
			customer_name text not null,
			reason text not null,
			scope text not null,
			severity text not null,
			block_orders boolean not null,
			block_dispatch boolean not null,
			status text not null,
			created_at text not null,
			released_at text not null,
			actor text not null
		);
		create index if not exists idx_cbmp_biz_customer_blacklists_customer on cbmp_biz_customer_blacklists (customer_id, status);

		create table if not exists cbmp_biz_customer_profiles (
			id bigint primary key,
			customer_id bigint not null,
			customer_name text not null,
			grade text not null,
			risk_level text not null,
			credit_score integer not null,
			tags text not null,
			status text not null,
			updated_at text not null,
			actor text not null
		);
		create index if not exists idx_cbmp_biz_customer_profiles_customer on cbmp_biz_customer_profiles (customer_id, grade, risk_level);

		create table if not exists cbmp_biz_customer_complaints (
			id bigint primary key,
			complaint_no text not null,
			customer_id bigint not null,
			project_id bigint not null,
			title text not null,
			level text not null,
			status text not null,
			owner text not null,
			sla_hours integer not null,
			due_at text not null,
			sla_status text not null,
			overdue_hours integer not null,
			created_at text not null,
			closed_at text not null
		);
		create index if not exists idx_cbmp_biz_customer_complaints_customer on cbmp_biz_customer_complaints (customer_id, status, level);

		create table if not exists cbmp_biz_contract_attachments (
			id bigint primary key,
			contract_id bigint not null,
			customer_id bigint not null,
			file_name text not null,
			file_type text not null,
			url text not null,
			checksum text not null,
			status text not null,
			uploaded_by text not null,
			uploaded_at text not null
		);
		create index if not exists idx_cbmp_biz_contract_attachments_contract on cbmp_biz_contract_attachments (contract_id, status);

		create table if not exists cbmp_biz_price_policies (
			id bigint primary key,
			customer_id bigint not null,
			project_id bigint not null,
			product_id bigint not null,
			customer_grade text not null,
			region text not null,
			min_quantity double precision not null,
			max_quantity double precision not null,
			floor_price double precision not null,
			sale_price double precision not null,
			promotion_name text not null,
			promotion_type text not null,
			promotion_value double precision not null,
			priority integer not null,
			tax_rate_id bigint not null,
			effective_from text not null,
			effective_to text not null,
			status text not null
		);
		create index if not exists idx_cbmp_biz_price_policies_scope on cbmp_biz_price_policies (customer_id, project_id, product_id, customer_grade, region, status);

		create table if not exists cbmp_biz_tax_rates (
			id bigint primary key,
			name text not null,
			rate double precision not null,
			scope text not null,
			status text not null
		);
		create index if not exists idx_cbmp_biz_tax_rates_scope on cbmp_biz_tax_rates (scope, status);

		create table if not exists cbmp_biz_orders (
			id bigint primary key,
			order_no text not null,
			customer_id bigint not null,
			project_id bigint not null,
			product_id bigint not null,
			site_id bigint not null,
			product_line text not null,
			plan_quantity double precision not null,
			signed_qty double precision not null,
			unit_price double precision not null,
			total_amount double precision not null,
			status text not null,
			risk_flag text not null,
			plan_time text not null,
			created_at text not null
		);
		create index if not exists idx_cbmp_biz_orders_customer on cbmp_biz_orders (customer_id, status);
		create index if not exists idx_cbmp_biz_orders_site on cbmp_biz_orders (site_id, plan_time);

		create table if not exists cbmp_biz_order_lines (
			id bigint primary key,
			order_id bigint not null,
			order_no text not null,
			seq integer not null,
			product_id bigint not null,
			product_line text not null,
			product_name text not null,
			strength_grade text not null,
			slump text not null,
			pouring_part text not null,
			quantity double precision not null,
			unit text not null,
			unit_price double precision not null,
			floor_price double precision not null,
			tax_rate double precision not null,
			amount double precision not null,
			price_source text not null,
			risk_flag text not null
		);
		create index if not exists idx_cbmp_biz_order_lines_order on cbmp_biz_order_lines (order_id, seq);
		create index if not exists idx_cbmp_biz_order_lines_product on cbmp_biz_order_lines (product_id, product_line);

		create table if not exists cbmp_biz_dispatch_orders (
			id bigint primary key,
			dispatch_no text not null,
			order_id bigint not null,
			vehicle_id bigint not null,
			driver_id bigint not null,
			site_id bigint not null,
			project_id bigint not null,
			line_id bigint not null,
			line_seq integer not null,
			product_id bigint not null,
			product_name text not null,
			plan_quantity double precision not null,
			loaded_qty double precision not null,
			signed_qty double precision not null,
			status text not null,
			exception text not null,
			created_at text not null,
			updated_at text not null
		);
		create index if not exists idx_cbmp_biz_dispatch_order on cbmp_biz_dispatch_orders (order_id, line_id, status);

		create table if not exists cbmp_biz_dispatch_schedules (
			id bigint primary key,
			schedule_no text not null,
			site_id bigint not null,
			vehicle_id bigint not null,
			driver_id bigint not null,
			carrier_id bigint not null,
			shift_date text not null,
			shift text not null,
			capacity_qty double precision not null,
			assigned_qty double precision not null,
			status text not null,
			created_at text not null,
			updated_at text not null
		);
		create index if not exists idx_cbmp_biz_dispatch_schedules_vehicle on cbmp_biz_dispatch_schedules (vehicle_id, shift_date, shift);

		create table if not exists cbmp_biz_transport_settlements (
			id bigint primary key,
			settlement_no text not null,
			carrier_id bigint not null,
			period text not null,
			trip_count integer not null,
			amount double precision not null,
			status text not null
		);
		create index if not exists idx_cbmp_biz_transport_settlement_carrier on cbmp_biz_transport_settlements (carrier_id, period, status);

		create table if not exists cbmp_biz_transport_settlement_items (
			id bigint primary key,
			settlement_id bigint not null,
			dispatch_id bigint not null,
			dispatch_no text not null,
			carrier_id bigint not null,
			vehicle_id bigint not null,
			driver_id bigint not null,
			quantity double precision not null,
			amount double precision not null,
			status text not null,
			created_at text not null
		);
		create index if not exists idx_cbmp_biz_transport_settlement_items_settlement on cbmp_biz_transport_settlement_items (settlement_id, status);

		create table if not exists cbmp_biz_mix_designs (
			id bigint primary key,
			product_id bigint not null,
			site_id bigint not null,
			parent_id bigint not null,
			code text not null,
			version text not null,
			strength_grade text not null,
			slump text not null,
			scope text not null,
			status text not null,
			is_current boolean not null,
			effective_from text not null,
			effective_to text not null,
			approved_by text not null,
			approved_at text not null,
			retired_at text not null,
			created_by text not null,
			created_at text not null,
			updated_at text not null
		);
		create index if not exists idx_cbmp_biz_mix_designs_product_site on cbmp_biz_mix_designs (product_id, site_id, status);

		create table if not exists cbmp_biz_mix_design_trial_runs (
			id bigint primary key,
			trial_no text not null,
			mix_design_id bigint not null,
			product_id bigint not null,
			site_id bigint not null,
			target_strength text not null,
			slump text not null,
			water double precision not null,
			sand_rate double precision not null,
			admixture_rate double precision not null,
			strength_7d double precision not null,
			strength_28d double precision not null,
			result text not null,
			conclusion text not null,
			tester text not null,
			tested_at text not null,
			created_at text not null
		);
		create index if not exists idx_cbmp_biz_mix_trials_design on cbmp_biz_mix_design_trial_runs (mix_design_id, result);

		create table if not exists cbmp_biz_laboratory_samples (
			id bigint primary key,
			sample_no text not null,
			source_type text not null,
			source_id bigint not null,
			site_id bigint not null,
			product_id bigint not null,
			material_id bigint not null,
			mix_design_id bigint not null,
			batch_id bigint not null,
			sample_type text not null,
			status text not null,
			result text not null,
			planned_test_at text not null,
			collected_at text not null,
			created_by text not null
		);
		create index if not exists idx_cbmp_biz_lab_samples_site on cbmp_biz_laboratory_samples (site_id, status, result);

		create table if not exists cbmp_biz_laboratory_tests (
			id bigint primary key,
			test_no text not null,
			sample_id bigint not null,
			equipment_id bigint not null,
			site_id bigint not null,
			test_type text not null,
			metric text not null,
			value double precision not null,
			unit text not null,
			result text not null,
			status text not null,
			tester text not null,
			tested_at text not null,
			reviewer text not null,
			reviewed_at text not null
		);
		create index if not exists idx_cbmp_biz_lab_tests_sample on cbmp_biz_laboratory_tests (sample_id, status, result);

		create table if not exists cbmp_biz_laboratory_equipment (
			id bigint primary key,
			equipment_no text not null,
			name text not null,
			site_id bigint not null,
			model text not null,
			serial_no text not null,
			status text not null,
			calibration_cycle_days integer not null,
			last_calibration_at text not null,
			next_calibration_at text not null,
			created_at text not null
		);
		create index if not exists idx_cbmp_biz_lab_equipment_site on cbmp_biz_laboratory_equipment (site_id, status);

		create table if not exists cbmp_biz_laboratory_calibrations (
			id bigint primary key,
			calibration_no text not null,
			equipment_id bigint not null,
			site_id bigint not null,
			result text not null,
			calibrated_at text not null,
			next_due_at text not null,
			certificate_no text not null,
			agency text not null,
			operator text not null
		);
		create index if not exists idx_cbmp_biz_lab_calibrations_equipment on cbmp_biz_laboratory_calibrations (equipment_id, result);

		create table if not exists cbmp_biz_quality_exceptions (
			id bigint primary key,
			exception_no text not null,
			source_type text not null,
			source_id bigint not null,
			site_id bigint not null,
			severity text not null,
			title text not null,
			status text not null,
			responsible text not null,
			created_at text not null,
			handled_at text not null,
			closed_by text not null
		);
		create index if not exists idx_cbmp_biz_quality_exceptions_site on cbmp_biz_quality_exceptions (site_id, status, severity);

		create table if not exists cbmp_biz_production_batches (
			id bigint primary key,
			batch_no text not null,
			task_id bigint not null,
			plan_id bigint not null,
			order_id bigint not null,
			site_id bigint not null,
			product_id bigint not null,
			quantity double precision not null,
			plant_code text not null,
			quality_status text not null,
			status text not null,
			started_at text not null,
			completed_at text not null
		);
		create index if not exists idx_cbmp_biz_batches_order on cbmp_biz_production_batches (order_id, status);

		create table if not exists cbmp_biz_scale_tickets (
			id bigint primary key,
			ticket_no text not null,
			ticket_type text not null,
			dispatch_id bigint not null,
			order_id bigint not null,
			site_id bigint not null,
			vehicle_id bigint not null,
			plate_no text not null,
			gross_weight double precision not null,
			tare_weight double precision not null,
			net_weight double precision not null,
			sign_status text not null,
			settlement_status text not null,
			status text not null,
			created_at text not null
		);
		create index if not exists idx_cbmp_biz_tickets_order on cbmp_biz_scale_tickets (order_id, status);
		create index if not exists idx_cbmp_biz_tickets_plate on cbmp_biz_scale_tickets (plate_no, created_at);

		create table if not exists cbmp_biz_scale_device_events (
			id bigint primary key,
			event_no text not null,
			device_id bigint not null,
			device_code text not null,
			ticket_id bigint not null,
			plate_no text not null,
			weight double precision not null,
			weight_type text not null,
			stable boolean not null,
			cheat_flag boolean not null,
			status text not null,
			received_at text not null
		);
		create index if not exists idx_cbmp_biz_scale_events_ticket on cbmp_biz_scale_device_events (ticket_id, status);

		create table if not exists cbmp_biz_delivery_sign_links (
			id bigint primary key,
			link_no text not null,
			dispatch_id bigint not null,
			ticket_id bigint not null,
			order_id bigint not null,
			line_id bigint not null,
			line_seq integer not null,
			product_id bigint not null,
			product_name text not null,
			customer_id bigint not null,
			project_id bigint not null,
			channel text not null,
			phone text not null,
			status text not null,
			sent_at text not null,
			expires_at text not null,
			used_at text not null,
			created_at text not null
		);
		create index if not exists idx_cbmp_biz_sign_links_dispatch on cbmp_biz_delivery_sign_links (dispatch_id, status);

		create table if not exists cbmp_biz_delivery_signs (
			id bigint primary key,
			sign_no text not null,
			dispatch_id bigint not null,
			ticket_id bigint not null,
			order_id bigint not null,
			line_id bigint not null,
			line_seq integer not null,
			product_id bigint not null,
			product_name text not null,
			customer_id bigint not null,
			project_id bigint not null,
			signed_qty double precision not null,
			signed_at text not null
		);
		create index if not exists idx_cbmp_biz_signs_order on cbmp_biz_delivery_signs (order_id, signed_at);

		create table if not exists cbmp_biz_delivery_sign_attachments (
			id bigint primary key,
			sign_id bigint not null,
			dispatch_id bigint not null,
			ticket_id bigint not null,
			file_name text not null,
			file_type text not null,
			url text not null,
			checksum text not null,
			uploaded_by text not null,
			uploaded_at text not null
		);
		create index if not exists idx_cbmp_biz_sign_attachments_sign on cbmp_biz_delivery_sign_attachments (sign_id, uploaded_at);

		create table if not exists cbmp_biz_sales_invoices (
			id bigint primary key,
			invoice_no text not null,
			statement_id bigint not null,
			customer_id bigint not null,
			amount double precision not null,
			tax_amount double precision not null,
			tax_control_no text not null,
			tax_status text not null,
			file_url text not null,
			status text not null,
			issued_at text not null,
			invoice_type text not null,
			invoice_category text not null,
			original_invoice_id bigint not null,
			red_letter_info_id bigint not null,
			red_letter_info_no text not null,
			red_reason text not null,
			red_at text not null
		);
		alter table cbmp_biz_sales_invoices add column if not exists tax_control_no text not null default '';
		alter table cbmp_biz_sales_invoices add column if not exists file_url text not null default '';
		alter table cbmp_biz_sales_invoices add column if not exists invoice_type text not null default 'blue';
		alter table cbmp_biz_sales_invoices add column if not exists invoice_category text not null default 'blue_vat_special';
		alter table cbmp_biz_sales_invoices add column if not exists original_invoice_id bigint not null default 0;
		alter table cbmp_biz_sales_invoices add column if not exists red_letter_info_id bigint not null default 0;
		alter table cbmp_biz_sales_invoices add column if not exists red_letter_info_no text not null default '';
		alter table cbmp_biz_sales_invoices add column if not exists red_reason text not null default '';
		alter table cbmp_biz_sales_invoices add column if not exists red_at text not null default '';
		create index if not exists idx_cbmp_biz_invoices_customer on cbmp_biz_sales_invoices (customer_id, status);

		create table if not exists cbmp_biz_red_letter_infos (
			id bigint primary key,
			info_no text not null,
			original_invoice_id bigint not null,
			original_invoice_no text not null,
			red_invoice_id bigint not null,
			customer_id bigint not null,
			amount double precision not null,
			tax_amount double precision not null,
			reason text not null,
			applicant text not null,
			status text not null,
			tax_control_no text not null,
			requested_at text not null,
			approved_by text not null,
			approved_at text not null,
			used_at text not null
		);
		create index if not exists idx_cbmp_biz_red_letters_original on cbmp_biz_red_letter_infos (original_invoice_id, status);
		create index if not exists idx_cbmp_biz_red_letters_customer on cbmp_biz_red_letter_infos (customer_id, status);

		create table if not exists cbmp_biz_tax_gateway_submissions (
			id bigint primary key,
			submission_no text not null,
			invoice_id bigint not null,
			invoice_no text not null,
			action text not null,
			provider text not null,
			endpoint text not null,
			request_id text not null,
			status text not null,
			tax_control_no text not null,
			file_url text not null,
			error text not null,
			attempt integer not null,
			duration_ms bigint not null,
			submitted_at text not null,
			completed_at text not null,
			actor text not null
		);
		alter table cbmp_biz_tax_gateway_submissions add column if not exists action text not null default 'issue';
		create index if not exists idx_cbmp_biz_tax_submissions_invoice on cbmp_biz_tax_gateway_submissions (invoice_id, status);
		create index if not exists idx_cbmp_biz_tax_submissions_request on cbmp_biz_tax_gateway_submissions (request_id);

		create table if not exists cbmp_biz_receivables (
			id bigint primary key,
			bill_no text not null,
			customer_id bigint not null,
			statement_id bigint not null,
			invoice_id bigint not null,
			amount double precision not null,
			received_amount double precision not null,
			due_date text not null,
			status text not null,
			created_at text not null
		);
		create index if not exists idx_cbmp_biz_receivables_customer on cbmp_biz_receivables (customer_id, status, due_date);

		create table if not exists cbmp_biz_payment_plans (
			id bigint primary key,
			plan_no text not null,
			receivable_id bigint not null,
			customer_id bigint not null,
			amount double precision not null,
			due_date text not null,
			method text not null,
			status text not null,
			created_at text not null,
			settled_at text not null
		);
		create index if not exists idx_cbmp_biz_payment_plans_customer on cbmp_biz_payment_plans (customer_id, status, due_date);

		create table if not exists cbmp_biz_collection_tasks (
			id bigint primary key,
			task_no text not null,
			receivable_id bigint not null,
			customer_id bigint not null,
			customer_name text not null,
			amount double precision not null,
			due_date text not null,
			overdue_days integer not null,
			level text not null,
			channel text not null,
			status text not null,
			template_id bigint not null,
			send_count integer not null,
			last_sent_at text not null,
			generated_at text not null,
			handled_at text not null
		);
		alter table cbmp_biz_collection_tasks add column if not exists template_id bigint not null default 0;
		alter table cbmp_biz_collection_tasks add column if not exists send_count integer not null default 0;
		alter table cbmp_biz_collection_tasks add column if not exists last_sent_at text not null default '';
		create index if not exists idx_cbmp_biz_collection_customer on cbmp_biz_collection_tasks (customer_id, status, level);

		create table if not exists cbmp_biz_collection_templates (
			id bigint primary key,
			code text not null,
			name text not null,
			level text not null,
			channel text not null,
			content text not null,
			enabled boolean not null,
			updated_at text not null
		);
		create index if not exists idx_cbmp_biz_collection_templates_level on cbmp_biz_collection_templates (level, channel, enabled);

		create table if not exists cbmp_biz_collection_dispatches (
			id bigint primary key,
			dispatch_no text not null,
			task_id bigint not null,
			template_id bigint not null,
			customer_id bigint not null,
			channel text not null,
			target text not null,
			content text not null,
			endpoint text not null,
			status text not null,
			error text not null,
			sent_at text not null,
			actor text not null
		);
		create index if not exists idx_cbmp_biz_collection_dispatch_customer on cbmp_biz_collection_dispatches (customer_id, status, sent_at);

		create table if not exists cbmp_biz_inventory_items (
			id bigint primary key,
			site_id bigint not null,
			warehouse text not null,
			silo text not null,
			material_id bigint not null,
			batch_no text not null,
			raw_receipt_id bigint not null,
			supplier_id bigint not null,
			quantity double precision not null,
			unit text not null,
			quality_status text not null,
			available_status text not null,
			updated_at text not null
		);
		create index if not exists idx_cbmp_biz_inventory_site_material on cbmp_biz_inventory_items (site_id, material_id, available_status);

		create table if not exists cbmp_biz_inventory_flows (
			id bigint primary key,
			flow_no text not null,
			site_id bigint not null,
			material_id bigint not null,
			source_type text not null,
			source_id bigint not null,
			direction text not null,
			quantity double precision not null,
			balance_qty double precision not null,
			created_at text not null
		);
		create index if not exists idx_cbmp_biz_inventory_flows_source on cbmp_biz_inventory_flows (source_type, source_id);

		create table if not exists cbmp_biz_vehicle_locations (
			id bigint primary key,
			vehicle_id bigint not null,
			plate_no text not null,
			driver_id bigint not null,
			dispatch_id bigint not null,
			device_id text not null,
			source_type text not null,
			longitude double precision not null,
			latitude double precision not null,
			speed double precision not null,
			online_status text not null,
			is_abnormal boolean not null,
			abnormal_type text not null,
			location_time text not null,
			receive_time text not null
		);
		create index if not exists idx_cbmp_biz_locations_vehicle_time on cbmp_biz_vehicle_locations (vehicle_id, receive_time);
		create index if not exists idx_cbmp_biz_locations_plate_time on cbmp_biz_vehicle_locations (plate_no, receive_time);

		create table if not exists cbmp_biz_latest_locations (
			plate_no text not null,
			vehicle_id bigint not null,
			longitude double precision not null,
			latitude double precision not null,
			speed double precision not null,
			direction double precision not null,
			online_status text not null,
			transport_status text not null,
			last_location_time text not null,
			current_order_id bigint not null,
			current_project_id bigint not null,
			current_site_id bigint not null,
			current_customer_id bigint not null
		);
		create index if not exists idx_cbmp_biz_latest_locations_plate on cbmp_biz_latest_locations (plate_no);

		create table if not exists cbmp_biz_device_protocol_frames (
			id bigint primary key,
			frame_no text not null,
			channel text not null,
			protocol text not null,
			device_no text not null,
			parsed_resource text not null,
			parsed_id bigint not null,
			status text not null,
			error text not null,
			received_at text not null,
			actor text not null
		);
		create index if not exists idx_cbmp_biz_frames_status on cbmp_biz_device_protocol_frames (status, received_at);

		create table if not exists cbmp_biz_projection_status (
			id text primary key,
			table_count integer not null,
			row_count integer not null,
			snapshot_checksum text not null,
			refreshed_at timestamptz not null default now()
		);
	`
}

func (s *PostgresStore) refreshBusinessProjections(ctx context.Context, tx pgx.Tx, data AppData, snapshotChecksum string) error {
	for _, table := range postgresProjectionTableNames {
		if _, err := tx.Exec(ctx, "delete from "+table); err != nil {
			return err
		}
	}
	inserts := businessProjectionInserts(data)
	var batch pgx.Batch
	for _, insert := range inserts {
		batch.Queue(projectionInsertQuery(insert.table, insert.columns), insert.values...)
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
	_, err := tx.Exec(ctx, `
		insert into cbmp_biz_projection_status (id, table_count, row_count, snapshot_checksum, refreshed_at)
		values ('default', $1, $2, $3, now())
		on conflict (id)
		do update set table_count = excluded.table_count, row_count = excluded.row_count, snapshot_checksum = excluded.snapshot_checksum, refreshed_at = now()
	`, len(postgresProjectionTableNames), len(inserts), snapshotChecksum)
	return err
}

func businessProjectionRowCount(data AppData) int {
	return len(businessProjectionInserts(data))
}

func businessProjectionInserts(data AppData) []businessProjectionInsert {
	inserts := make([]businessProjectionInsert, 0)
	for _, item := range data.CustomerContacts {
		inserts = append(inserts, projection("cbmp_biz_customer_contacts",
			[]string{"id", "customer_id", "name", "phone", "role", "is_default", "status"},
			item.ID, item.CustomerID, item.Name, item.Phone, item.Role, item.IsDefault, item.Status,
		))
	}
	for _, item := range data.CustomerBlacklists {
		inserts = append(inserts, projection("cbmp_biz_customer_blacklists",
			[]string{"id", "customer_id", "customer_name", "reason", "scope", "severity", "block_orders", "block_dispatch", "status", "created_at", "released_at", "actor"},
			item.ID, item.CustomerID, item.CustomerName, item.Reason, item.Scope, item.Severity, item.BlockOrders, item.BlockDispatch, item.Status, item.CreatedAt, item.ReleasedAt, item.Actor,
		))
	}
	for _, item := range data.CustomerProfiles {
		inserts = append(inserts, projection("cbmp_biz_customer_profiles",
			[]string{"id", "customer_id", "customer_name", "grade", "risk_level", "credit_score", "tags", "status", "updated_at", "actor"},
			item.ID, item.CustomerID, item.CustomerName, item.Grade, item.RiskLevel, item.CreditScore, strings.Join(item.Tags, ","), item.Status, item.UpdatedAt, item.Actor,
		))
	}
	for _, item := range data.CustomerComplaints {
		inserts = append(inserts, projection("cbmp_biz_customer_complaints",
			[]string{"id", "complaint_no", "customer_id", "project_id", "title", "level", "status", "owner", "sla_hours", "due_at", "sla_status", "overdue_hours", "created_at", "closed_at"},
			item.ID, item.ComplaintNo, item.CustomerID, item.ProjectID, item.Title, item.Level, item.Status, item.Owner, item.SLAHours, item.DueAt, item.SLAStatus, item.OverdueHours, item.CreatedAt, item.ClosedAt,
		))
	}
	for _, item := range data.ContractAttachments {
		inserts = append(inserts, projection("cbmp_biz_contract_attachments",
			[]string{"id", "contract_id", "customer_id", "file_name", "file_type", "url", "checksum", "status", "uploaded_by", "uploaded_at"},
			item.ID, item.ContractID, item.CustomerID, item.FileName, item.FileType, item.URL, item.Checksum, item.Status, item.UploadedBy, item.UploadedAt,
		))
	}
	for _, item := range data.PricePolicies {
		inserts = append(inserts, projection("cbmp_biz_price_policies",
			[]string{"id", "customer_id", "project_id", "product_id", "customer_grade", "region", "min_quantity", "max_quantity", "floor_price", "sale_price", "promotion_name", "promotion_type", "promotion_value", "priority", "tax_rate_id", "effective_from", "effective_to", "status"},
			item.ID, item.CustomerID, item.ProjectID, item.ProductID, item.CustomerGrade, item.Region, item.MinQuantity, item.MaxQuantity, item.FloorPrice, item.SalePrice, item.PromotionName, item.PromotionType, item.PromotionValue, item.Priority, item.TaxRateID, item.EffectiveFrom, item.EffectiveTo, item.Status,
		))
	}
	for _, item := range data.TaxRates {
		inserts = append(inserts, projection("cbmp_biz_tax_rates",
			[]string{"id", "name", "rate", "scope", "status"},
			item.ID, item.Name, item.Rate, item.Scope, item.Status,
		))
	}
	for _, item := range data.Orders {
		inserts = append(inserts, projection("cbmp_biz_orders",
			[]string{"id", "order_no", "customer_id", "project_id", "product_id", "site_id", "product_line", "plan_quantity", "signed_qty", "unit_price", "total_amount", "status", "risk_flag", "plan_time", "created_at"},
			item.ID, item.OrderNo, item.CustomerID, item.ProjectID, item.ProductID, item.SiteID, item.ProductLine, item.PlanQuantity, item.SignedQty, item.UnitPrice, orderTotalAmount(item), item.Status, item.RiskFlag, item.PlanTime, item.CreatedAt,
		))
		for _, line := range orderLines(item) {
			inserts = append(inserts, projection("cbmp_biz_order_lines",
				[]string{"id", "order_id", "order_no", "seq", "product_id", "product_line", "product_name", "strength_grade", "slump", "pouring_part", "quantity", "unit", "unit_price", "floor_price", "tax_rate", "amount", "price_source", "risk_flag"},
				line.ID, item.ID, item.OrderNo, line.Seq, line.ProductID, line.ProductLine, line.ProductName, line.StrengthGrade, line.Slump, line.PouringPart, line.Quantity, line.Unit, line.UnitPrice, line.FloorPrice, line.TaxRate, line.Amount, line.PriceSource, line.RiskFlag,
			))
		}
	}
	for _, item := range data.DispatchOrders {
		inserts = append(inserts, projection("cbmp_biz_dispatch_orders",
			[]string{"id", "dispatch_no", "order_id", "vehicle_id", "driver_id", "site_id", "project_id", "line_id", "line_seq", "product_id", "product_name", "plan_quantity", "loaded_qty", "signed_qty", "status", "exception", "created_at", "updated_at"},
			item.ID, item.DispatchNo, item.OrderID, item.VehicleID, item.DriverID, item.SiteID, item.ProjectID, item.LineID, item.LineSeq, item.ProductID, item.ProductName, item.PlanQuantity, item.LoadedQty, item.SignedQty, item.Status, item.Exception, item.CreatedAt, item.UpdatedAt,
		))
	}
	for _, item := range data.DispatchSchedules {
		inserts = append(inserts, projection("cbmp_biz_dispatch_schedules",
			[]string{"id", "schedule_no", "site_id", "vehicle_id", "driver_id", "carrier_id", "shift_date", "shift", "capacity_qty", "assigned_qty", "status", "created_at", "updated_at"},
			item.ID, item.ScheduleNo, item.SiteID, item.VehicleID, item.DriverID, item.CarrierID, item.ShiftDate, item.Shift, item.CapacityQty, item.AssignedQty, item.Status, item.CreatedAt, item.UpdatedAt,
		))
	}
	for _, item := range data.TransportSettlements {
		inserts = append(inserts, projection("cbmp_biz_transport_settlements",
			[]string{"id", "settlement_no", "carrier_id", "period", "trip_count", "amount", "status"},
			item.ID, item.SettlementNo, item.CarrierID, item.Period, item.TripCount, item.Amount, item.Status,
		))
	}
	for _, item := range data.TransportSettlementItems {
		inserts = append(inserts, projection("cbmp_biz_transport_settlement_items",
			[]string{"id", "settlement_id", "dispatch_id", "dispatch_no", "carrier_id", "vehicle_id", "driver_id", "quantity", "amount", "status", "created_at"},
			item.ID, item.SettlementID, item.DispatchID, item.DispatchNo, item.CarrierID, item.VehicleID, item.DriverID, item.Quantity, item.Amount, item.Status, item.CreatedAt,
		))
	}
	for _, item := range data.MixDesigns {
		inserts = append(inserts, projection("cbmp_biz_mix_designs",
			[]string{"id", "product_id", "site_id", "parent_id", "code", "version", "strength_grade", "slump", "scope", "status", "is_current", "effective_from", "effective_to", "approved_by", "approved_at", "retired_at", "created_by", "created_at", "updated_at"},
			item.ID, item.ProductID, item.SiteID, item.ParentID, item.Code, item.Version, item.StrengthGrade, item.Slump, item.Scope, item.Status, item.IsCurrent, item.EffectiveFrom, item.EffectiveTo, item.ApprovedBy, item.ApprovedAt, item.RetiredAt, item.CreatedBy, item.CreatedAt, item.UpdatedAt,
		))
	}
	for _, item := range data.MixDesignTrialRuns {
		inserts = append(inserts, projection("cbmp_biz_mix_design_trial_runs",
			[]string{"id", "trial_no", "mix_design_id", "product_id", "site_id", "target_strength", "slump", "water", "sand_rate", "admixture_rate", "strength_7d", "strength_28d", "result", "conclusion", "tester", "tested_at", "created_at"},
			item.ID, item.TrialNo, item.MixDesignID, item.ProductID, item.SiteID, item.TargetStrength, item.Slump, item.Water, item.SandRate, item.AdmixtureRate, item.Strength7d, item.Strength28d, item.Result, item.Conclusion, item.Tester, item.TestedAt, item.CreatedAt,
		))
	}
	for _, item := range data.LaboratorySamples {
		inserts = append(inserts, projection("cbmp_biz_laboratory_samples",
			[]string{"id", "sample_no", "source_type", "source_id", "site_id", "product_id", "material_id", "mix_design_id", "batch_id", "sample_type", "status", "result", "planned_test_at", "collected_at", "created_by"},
			item.ID, item.SampleNo, item.SourceType, item.SourceID, item.SiteID, item.ProductID, item.MaterialID, item.MixDesignID, item.BatchID, item.SampleType, item.Status, item.Result, item.PlannedTestAt, item.CollectedAt, item.CreatedBy,
		))
	}
	for _, item := range data.LaboratoryTests {
		inserts = append(inserts, projection("cbmp_biz_laboratory_tests",
			[]string{"id", "test_no", "sample_id", "equipment_id", "site_id", "test_type", "metric", "value", "unit", "result", "status", "tester", "tested_at", "reviewer", "reviewed_at"},
			item.ID, item.TestNo, item.SampleID, item.EquipmentID, item.SiteID, item.TestType, item.Metric, item.Value, item.Unit, item.Result, item.Status, item.Tester, item.TestedAt, item.Reviewer, item.ReviewedAt,
		))
	}
	for _, item := range data.LaboratoryEquipment {
		inserts = append(inserts, projection("cbmp_biz_laboratory_equipment",
			[]string{"id", "equipment_no", "name", "site_id", "model", "serial_no", "status", "calibration_cycle_days", "last_calibration_at", "next_calibration_at", "created_at"},
			item.ID, item.EquipmentNo, item.Name, item.SiteID, item.Model, item.SerialNo, item.Status, item.CalibrationCycleDays, item.LastCalibrationAt, item.NextCalibrationAt, item.CreatedAt,
		))
	}
	for _, item := range data.LaboratoryCalibrations {
		inserts = append(inserts, projection("cbmp_biz_laboratory_calibrations",
			[]string{"id", "calibration_no", "equipment_id", "site_id", "result", "calibrated_at", "next_due_at", "certificate_no", "agency", "operator"},
			item.ID, item.CalibrationNo, item.EquipmentID, item.SiteID, item.Result, item.CalibratedAt, item.NextDueAt, item.CertificateNo, item.Agency, item.Operator,
		))
	}
	for _, item := range data.QualityExceptions {
		inserts = append(inserts, projection("cbmp_biz_quality_exceptions",
			[]string{"id", "exception_no", "source_type", "source_id", "site_id", "severity", "title", "status", "responsible", "created_at", "handled_at", "closed_by"},
			item.ID, item.ExceptionNo, item.SourceType, item.SourceID, item.SiteID, item.Severity, item.Title, item.Status, item.Responsible, item.CreatedAt, item.HandledAt, item.ClosedBy,
		))
	}
	for _, item := range data.ProductionBatches {
		inserts = append(inserts, projection("cbmp_biz_production_batches",
			[]string{"id", "batch_no", "task_id", "plan_id", "order_id", "site_id", "product_id", "quantity", "plant_code", "quality_status", "status", "started_at", "completed_at"},
			item.ID, item.BatchNo, item.TaskID, item.PlanID, item.OrderID, item.SiteID, item.ProductID, item.Quantity, item.PlantCode, item.QualityStatus, item.Status, item.StartedAt, item.CompletedAt,
		))
	}
	for _, item := range data.ScaleTickets {
		inserts = append(inserts, projection("cbmp_biz_scale_tickets",
			[]string{"id", "ticket_no", "ticket_type", "dispatch_id", "order_id", "site_id", "vehicle_id", "plate_no", "gross_weight", "tare_weight", "net_weight", "sign_status", "settlement_status", "status", "created_at"},
			item.ID, item.TicketNo, item.TicketType, item.DispatchID, item.OrderID, item.SiteID, item.VehicleID, item.PlateNo, item.GrossWeight, item.TareWeight, item.NetWeight, item.SignStatus, item.SettlementStatus, item.Status, item.CreatedAt,
		))
	}
	for _, item := range data.ScaleDeviceEvents {
		inserts = append(inserts, projection("cbmp_biz_scale_device_events",
			[]string{"id", "event_no", "device_id", "device_code", "ticket_id", "plate_no", "weight", "weight_type", "stable", "cheat_flag", "status", "received_at"},
			item.ID, item.EventNo, item.DeviceID, item.DeviceCode, item.TicketID, item.PlateNo, item.Weight, item.WeightType, item.Stable, item.CheatFlag, item.Status, item.ReceivedAt,
		))
	}
	for _, item := range data.DeliverySignLinks {
		inserts = append(inserts, projection("cbmp_biz_delivery_sign_links",
			[]string{"id", "link_no", "dispatch_id", "ticket_id", "order_id", "line_id", "line_seq", "product_id", "product_name", "customer_id", "project_id", "channel", "phone", "status", "sent_at", "expires_at", "used_at", "created_at"},
			item.ID, item.LinkNo, item.DispatchID, item.TicketID, item.OrderID, item.LineID, item.LineSeq, item.ProductID, item.ProductName, item.CustomerID, item.ProjectID, item.Channel, item.Phone, item.Status, item.SentAt, item.ExpiresAt, item.UsedAt, item.CreatedAt,
		))
	}
	for _, item := range data.DeliverySigns {
		inserts = append(inserts, projection("cbmp_biz_delivery_signs",
			[]string{"id", "sign_no", "dispatch_id", "ticket_id", "order_id", "line_id", "line_seq", "product_id", "product_name", "customer_id", "project_id", "signed_qty", "signed_at"},
			item.ID, item.SignNo, item.DispatchID, item.TicketID, item.OrderID, item.LineID, item.LineSeq, item.ProductID, item.ProductName, item.CustomerID, item.ProjectID, item.SignedQty, item.SignedAt,
		))
	}
	for _, item := range data.DeliverySignAttachments {
		inserts = append(inserts, projection("cbmp_biz_delivery_sign_attachments",
			[]string{"id", "sign_id", "dispatch_id", "ticket_id", "file_name", "file_type", "url", "checksum", "uploaded_by", "uploaded_at"},
			item.ID, item.SignID, item.DispatchID, item.TicketID, item.FileName, item.FileType, item.URL, item.Checksum, item.UploadedBy, item.UploadedAt,
		))
	}
	for _, item := range data.SalesInvoices {
		inserts = append(inserts, projection("cbmp_biz_sales_invoices",
			[]string{"id", "invoice_no", "statement_id", "customer_id", "amount", "tax_amount", "tax_control_no", "tax_status", "file_url", "status", "issued_at", "invoice_type", "invoice_category", "original_invoice_id", "red_letter_info_id", "red_letter_info_no", "red_reason", "red_at"},
			item.ID, item.InvoiceNo, item.StatementID, item.CustomerID, item.Amount, item.TaxAmount, item.TaxControlNo, item.TaxStatus, item.FileURL, item.Status, item.IssuedAt, fallback(item.InvoiceType, "blue"), fallback(item.InvoiceCategory, "blue_vat_special"), item.OriginalInvoiceID, item.RedLetterInfoID, item.RedLetterInfoNo, item.RedReason, item.RedAt,
		))
	}
	for _, item := range data.RedLetterInfos {
		inserts = append(inserts, projection("cbmp_biz_red_letter_infos",
			[]string{"id", "info_no", "original_invoice_id", "original_invoice_no", "red_invoice_id", "customer_id", "amount", "tax_amount", "reason", "applicant", "status", "tax_control_no", "requested_at", "approved_by", "approved_at", "used_at"},
			item.ID, item.InfoNo, item.OriginalInvoiceID, item.OriginalInvoiceNo, item.RedInvoiceID, item.CustomerID, item.Amount, item.TaxAmount, item.Reason, item.Applicant, item.Status, item.TaxControlNo, item.RequestedAt, item.ApprovedBy, item.ApprovedAt, item.UsedAt,
		))
	}
	for _, item := range data.TaxGatewaySubmissions {
		inserts = append(inserts, projection("cbmp_biz_tax_gateway_submissions",
			[]string{"id", "submission_no", "invoice_id", "invoice_no", "action", "provider", "endpoint", "request_id", "status", "tax_control_no", "file_url", "error", "attempt", "duration_ms", "submitted_at", "completed_at", "actor"},
			item.ID, item.SubmissionNo, item.InvoiceID, item.InvoiceNo, fallback(item.Action, "issue"), item.Provider, item.Endpoint, item.RequestID, item.Status, item.TaxControlNo, item.FileURL, item.Error, item.Attempt, item.DurationMs, item.SubmittedAt, item.CompletedAt, item.Actor,
		))
	}
	for _, item := range data.Receivables {
		inserts = append(inserts, projection("cbmp_biz_receivables",
			[]string{"id", "bill_no", "customer_id", "statement_id", "invoice_id", "amount", "received_amount", "due_date", "status", "created_at"},
			item.ID, item.BillNo, item.CustomerID, item.StatementID, item.InvoiceID, item.Amount, item.ReceivedAmount, item.DueDate, item.Status, item.CreatedAt,
		))
	}
	for _, item := range data.PaymentPlans {
		inserts = append(inserts, projection("cbmp_biz_payment_plans",
			[]string{"id", "plan_no", "receivable_id", "customer_id", "amount", "due_date", "method", "status", "created_at", "settled_at"},
			item.ID, item.PlanNo, item.ReceivableID, item.CustomerID, item.Amount, item.DueDate, item.Method, item.Status, item.CreatedAt, item.SettledAt,
		))
	}
	for _, item := range data.CollectionTasks {
		inserts = append(inserts, projection("cbmp_biz_collection_tasks",
			[]string{"id", "task_no", "receivable_id", "customer_id", "customer_name", "amount", "due_date", "overdue_days", "level", "channel", "status", "template_id", "send_count", "last_sent_at", "generated_at", "handled_at"},
			item.ID, item.TaskNo, item.ReceivableID, item.CustomerID, item.CustomerName, item.Amount, item.DueDate, item.OverdueDays, item.Level, item.Channel, item.Status, item.TemplateID, item.SendCount, item.LastSentAt, item.GeneratedAt, item.HandledAt,
		))
	}
	for _, item := range data.CollectionTemplates {
		inserts = append(inserts, projection("cbmp_biz_collection_templates",
			[]string{"id", "code", "name", "level", "channel", "content", "enabled", "updated_at"},
			item.ID, item.Code, item.Name, item.Level, item.Channel, item.Content, item.Enabled, item.UpdatedAt,
		))
	}
	for _, item := range data.CollectionDispatches {
		inserts = append(inserts, projection("cbmp_biz_collection_dispatches",
			[]string{"id", "dispatch_no", "task_id", "template_id", "customer_id", "channel", "target", "content", "endpoint", "status", "error", "sent_at", "actor"},
			item.ID, item.DispatchNo, item.TaskID, item.TemplateID, item.CustomerID, item.Channel, item.Target, item.Content, item.Endpoint, item.Status, item.Error, item.SentAt, item.Actor,
		))
	}
	for _, item := range data.Inventory {
		inserts = append(inserts, projection("cbmp_biz_inventory_items",
			[]string{"id", "site_id", "warehouse", "silo", "material_id", "batch_no", "raw_receipt_id", "supplier_id", "quantity", "unit", "quality_status", "available_status", "updated_at"},
			item.ID, item.SiteID, item.Warehouse, item.Silo, item.MaterialID, item.BatchNo, item.RawReceiptID, item.SupplierID, item.Quantity, item.Unit, item.QualityStatus, item.AvailableStatus, item.UpdatedAt,
		))
	}
	for _, item := range data.InventoryFlows {
		inserts = append(inserts, projection("cbmp_biz_inventory_flows",
			[]string{"id", "flow_no", "site_id", "material_id", "source_type", "source_id", "direction", "quantity", "balance_qty", "created_at"},
			item.ID, item.FlowNo, item.SiteID, item.MaterialID, item.SourceType, item.SourceID, item.Direction, item.Quantity, item.BalanceQty, item.CreatedAt,
		))
	}
	for _, item := range data.Locations {
		inserts = append(inserts, projection("cbmp_biz_vehicle_locations",
			[]string{"id", "vehicle_id", "plate_no", "driver_id", "dispatch_id", "device_id", "source_type", "longitude", "latitude", "speed", "online_status", "is_abnormal", "abnormal_type", "location_time", "receive_time"},
			item.ID, item.VehicleID, item.PlateNo, item.DriverID, item.DispatchID, item.DeviceID, item.SourceType, item.Longitude, item.Latitude, item.Speed, item.OnlineStatus, item.IsAbnormal, item.AbnormalType, item.LocationTime, item.ReceiveTime,
		))
	}
	for _, item := range data.LatestLocations {
		inserts = append(inserts, projection("cbmp_biz_latest_locations",
			[]string{"plate_no", "vehicle_id", "longitude", "latitude", "speed", "direction", "online_status", "transport_status", "last_location_time", "current_order_id", "current_project_id", "current_site_id", "current_customer_id"},
			item.PlateNo, item.VehicleID, item.Longitude, item.Latitude, item.Speed, item.Direction, item.OnlineStatus, item.TransportStatus, item.LastLocationTime, item.CurrentOrderID, item.CurrentProjectID, item.CurrentSiteID, item.CurrentCustomerID,
		))
	}
	for _, item := range data.DeviceProtocolFrames {
		inserts = append(inserts, projection("cbmp_biz_device_protocol_frames",
			[]string{"id", "frame_no", "channel", "protocol", "device_no", "parsed_resource", "parsed_id", "status", "error", "received_at", "actor"},
			item.ID, item.FrameNo, item.Channel, item.Protocol, item.DeviceNo, item.ParsedResource, item.ParsedID, item.Status, item.Error, item.ReceivedAt, item.Actor,
		))
	}
	return inserts
}

func projection(table string, columns []string, values ...interface{}) businessProjectionInsert {
	return businessProjectionInsert{table: table, columns: columns, values: values}
}

func projectionInsertQuery(table string, columns []string) string {
	return fmt.Sprintf("insert into %s (%s) values (%s)", table, strings.Join(columns, ", "), postgresPlaceholders(len(columns)))
}

func postgresPlaceholders(count int) string {
	parts := make([]string, count)
	for i := 0; i < count; i++ {
		parts[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(parts, ", ")
}
