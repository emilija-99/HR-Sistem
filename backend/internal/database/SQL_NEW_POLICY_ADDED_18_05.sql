-- user login
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active boolean default true
);

-- roles for user action
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);
INSERT INTO roles VALUES (1,'user');
INSERT INTO roles VALUES (2,'admin');
INSERT INTO roles VALUES (3,'approver');

-- premissions - whitch user can perform action
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);
INSERT into permissions values (1, 'absence.read');
INSERT into permissions values (2, 'absence.create');
INSERT into permissions values (3, 'absence.approve');
INSERT into permissions values (4, 'absence.reject');
INSERT into permissions values (5, 'absence.delete');
INSERT into permissions values (6, 'absence.cancel');
INSERT into permissions values (7, 'user.read');
INSERT into permissions values (8, 'user.edit');
INSERT into permissions values (9, 'user.delete');


-- stores user and roles groups
CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role_id INT REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);
-- which role has which premission
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- user can manage (read,create,cancel) only own absences
insert into role_permissions values (1,1);
insert into role_permissions values (1,2);
insert into role_permissions values (1,6);

-- admin is superadmin, have all premissions
insert into role_permissions values (2,1);
insert into role_permissions values (2,2);
insert into role_permissions values (2,3);
insert into role_permissions values (2,4);
insert into role_permissions values (2,5);
insert into role_permissions values (2,6);

-- approver can be team lead, project manager, superviors
-- can approve, reject, cancel, create for user / employee
insert into role_permissions values (3,1);
insert into role_permissions values (3,2);
insert into role_permissions values (3,3);
insert into role_permissions values (3,4);
insert into role_permissions values (3,6);

-- user - me (user page), can edit, read
insert into role_permissions values (1,7);
insert into role_permissions values (1,8);

-- admin - me - superuser
insert into role_permissions values (2,7);
insert into role_permissions values (2,8);
insert into role_permissions values (2,9);

--approver - me - can view, edit
insert into role_permissions values (3,7);
insert into role_permissions values (3,8);

-- status type - enum
create type STATUS as enum ('ACTIVE','INACTIVE','DELETED','MISC')

-- token refresh for storing user login
create table refresh_tokens(
  id UUID primary key,
  user_id BIGINT not null references users(id) on delete cascade,
  token_hash text not null,
  expires_at timestamp not null,
  created_at timestamp not null default now(),
  revoked boolean not null default false,
  user_agent text,
  ip_address text
);

create index idx_refresh_user on refresh_tokens(user_id)

-- get all user premissions

SELECT DISTINCT p.id, p.name
FROM permissions p
WHERE p.id IN (
    SELECT rp.permission_id
    FROM role_permissions rp
    WHERE rp.role_id = (
        SELECT r.id
        FROM roles r
        WHERE r.name = 'user'
    )
);
ORDER BY p.id;

-- absence types has status, means that all data is
-- always available and can be deleted, inactive, active
-- status of abstence type can be changed
-- defines what kinds of absence exists
CREATE TABLE IF NOT EXISTS absence_types (
    id INT PRIMARY KEY,
    type_name VARCHAR(40) NOT NULL,
    code VARCHAR(40) NOT NULL UNIQUE,
    is_paid BOOLEAN NOT NULL,
    status STATUS not null default 'ACTIVE',
    UNIQUE(code)
);

insert
  into
  absence_types
values (
  1,
    'Vacation',
    'VACATION',
  true
    );

INSERT INTO absence_types  VALUES(
  2,
  'Parental Leave',
  'PARENTAL',
  true);

INSERT INTO absence_types VALUES(
  3,
  'Sick Leave',
  'SICK',
  true);
INSERT INTO absence_types VALUES(
  4,
  'Training Leave',
  'TRAINING',
  true);

insert
  into
  absence_types
values(5, 'Disability Leave', 'DISABILITY', true)

INSERT INTO absence_types VALUES(
  6,
  'Personal Leave',
  'PERSONAL',
  true);

-- this represent state machine for every absence
CREATE TYPE absence_status AS ENUM('PENDING','APPROVED', 'REJECTED', 'CANCELLED','DRAFT');

--!!! absence_types → leave_policies → employee_leave_entitlements → ledger scheduler → leave_balance_ledger !!!--
--Absence type = category (Vacation, Sick…)
--Policy = rules for that category
--Entitlement = which employee uses which policy and when
--Scheduler reads entitlement → creates ledger ROWS


-- it allowes modeling of vacation, sickleave,parent leavem future rules without need for chaning schema
-- for every absence there is policy that represents system behavior for compute days for leave / absence
-- YEARLY_GRANT: employee gets 24 days every year
-- MATERNITY_GRANT: employee gets 110 days
-- UNLIMITED: employee is sick or there is not limit for days of absence
-- MONTLY_GRANT: there is limited days leave per month for absence type
-- MISC: possibility for other calculations
-- Employee → Absence Type → Policy: Explain which rules apply to this employee right now
CREATE TYPE GRANT_POLICY AS ENUM ('YEARLY_GRANT', 'MATERNITY_GRANT','MONTLY_GRANT','MISC');

CREATE TABLE leave_policies (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  grant_policy GRANT_POLICY NOT NULL,
  days_per_period NUMERIC(6,2),
  period_months INT,
  allow_carry_over BOOLEAN DEFAULT FALSE,
  carry_over_max_days NUMERIC(6,2),
  carry_over_expiry_month INT,
  carry_over_expiry_day INT,
  requires_balance BOOLEAN DEFAULT TRUE
);

-- 20 days per year, carry-over allowed until 30 June.
INSERT INTO leave_policies
(name, grant_policy, days_per_period, period_months,
 allow_carry_over, carry_over_max_days,
 carry_over_expiry_month, carry_over_expiry_day, requires_balance)
VALUES
(
 'Vacation standard',
 'YEARLY_GRANT',
 20,
 12,
 TRUE,
 10,
 6,   -- June
 30,
 TRUE
);
-- No days_per_period
--requires_balance = FALSE → skip ledger check
INSERT INTO leave_policies
(name, grant_policy, requires_balance)
VALUES
(
 'Sick unlimited',
 'MISC',
 FALSE
);
-- 365 once
INSERT INTO leave_policies
(name, grant_policy, days_per_period, requires_balance)
VALUES
(
 'Parental leave',
 'MATERNITY_GRANT',
 365,
 TRUE
);

-- employee policy assigment
-- this table connect policy to employe and per time period
-- table policy history for each employee
-- which leave rules applied to this employee on a given date, becuase
-- rules are for developers that works +5 days +1day per year
-- part time - hr adjust
-- country change, different law
-- promotion, next year has more days
-- Employee → Absence type → Policy → Ledger generation
CREATE TABLE employee_leave_policy(
  id BIGSERIAL PRIMARY KEY,
  employee_id BIGINT NOT NULL REFERENCES employees(id),
  absence_type_id INT NOT NULL REFERENCES absence_types(id),
  policy_id BIGINT NOT NULL REFERENCES leave_policies(id),
  valid_from DATE NOT NULL,
  valid_to DATE,
  UNIQUE(employee_id, absence_type_id, valid_from)
);

-- REPRESENTATION OF POLICY HISTORY
-- From 10 March 2024, this employee follows Vacation Standard policy.
-- Employee joined March 10, 2024.
INSERT INTO employee_leave_policy
(employee_id, absence_type_id, policy_id, valid_from)
VALUES (1, 1, 1, '2024-03-10');

UPDATE employee_leave_policy
SET valid_to = '2026-03-31'
WHERE employee_id = 1
AND absence_type_id = 1
AND valid_to IS NULL;

-- Employee becomes Senior → 25 days policy.
INSERT INTO employee_leave_policy
(employee_id, absence_type_id, policy_id, valid_from)
VALUES (1, 1, 2, '2026-04-01');


-- ACCURAL - days granted by policy [GRANT_POLICY] (+),
-- CARRY_OVER - days moved from previous year (+)
-- CONSUMED - approved absences (-)
-- CANCELED - returened days after cancel (+)
-- EXPIRATION - days expired JUNE 30 (-)
-- MANUAL_ADJUST - HR can (admin) corrent this
CREATE TYPE LEAVE_TYPE AS ENUM ('ACCURAL','CARRY_OVER','CONSUMED','CANCELED','MANUAL_ADJUST','EXPIRATION')

-- table represent calculation for remaning days
-- balance is computed: SUM(days where expires_at >= today)
-- reference_id pointes to absence_requests.id for CONSUMED, CANCELLED
-- Why did employee lose 5 days?
--  find ledger row
--  reference absence request
--  see approval history

CREATE TABLE leave_balance(
  id BIGSERIAL PRIMARY KEY,
  employee_id BIGINT NOT NULL REFERENCES employees(id),
  absence_type_id INT NOT NULL REFERENCES absence_types(id),
  entry_type_leave LEAVE_TYPE NOT NULL,
  days NUMERIC(6,2) NOT NULL,
  accrual_year INT NOT NULL,
  expires_at DATE,
  reference_id BIGINT,
  created_at timestamp DEFAULT now()
);

INSERT INTO public.departments (id, "name", description, "status", head_employee_id)
VALUES(0, 'GUI', 'Frontend developers', 'ACTIVE'::status, 0);

INSERT INTO public.positions (id, departments_id, title, "level", description, "status")
VALUES(nextval('positions_id_seq'::regclass), 0, 'Junior Fronend developer', 'JUNIOR'::position_level, 'Junior II', 'ACTIVE'::character varying);

INSERT INTO public.employees (id, user_id, first_name, last_name, phone_number, email_address, private_email, street, hire_date, position_id)
VALUES(nextval('employees_id_seq'::regclass), 1, 'Emilija', 'Ristic', '+38169747795', 'emilijaristic@hrsystem.com', 'emaristic53@gmail.com', 'BB', '2023-02-02', 1);

-- employee gets 20 days for 2026
INSERT INTO public.leave_balance
(employee_id, absence_type_id, entry_type_leave, days, accrual_year, expires_at)
VALUES (1, 1, 'ACCURAL', 20, 2025, '2026-06-30');
-- employee require 5 days
INSERT INTO public.leave_balance
(employee_id, absence_type_id, entry_type_leave, days, accrual_year, reference_id)
VALUES (1, 1, 'CONSUMED', -5, 2025, 55);

-- carry over - inser example
--INSERT INTO public.leave_balance
--(employee_id, absence_type_id, entry_type, days, accrual_year, expires_at)
--VALUES (10, 1, 'CARRY_OVER', 12, 2025, '2025-06-30');

-- carry over - expired
--INSERT INTO public.leave_balance
--(employee_id, absence_type_id, entry_type, days, accrual_year)
--VALUES (10, 1, 'EXPIRATION', -4, 2025);

-- AVAILABLE DAYS QUERY > This is the current balance.
SELECT COALESCE(SUM(days),0) AS available_days
FROM public.leave_balance
WHERE employee_id = 1
AND absence_type_id = 1
AND (expires_at IS NULL OR expires_at >= CURRENT_DATE);

-- TOTAL DAYS USED:
SELECT ABS(SUM(days))
FROM public.leave_balance
WHERE employee_id = 1
AND absence_type_id = 1
AND entry_type_leave = 'CONSUMED';

-- TOTAL EVER DAYS GRANTED: FIX TYPOOOO
SELECT SUM(days)
FROM public.leave_balance
WHERE employee_id = 1
AND absence_type_id = 1
AND entry_type_leave IN ('ACCRUAL','CARRY_OVER');

SELECT accrual_year, expires_at, SUM(days) AS remaining
FROM public.leave_balance
WHERE employee_id = 1
AND absence_type_id = 1
GROUP BY accrual_year, expires_at
HAVING SUM(days) > 0
ORDER BY expires_at;

CREATE TABLE absence_requests (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(id),
    absence_type_id INT NOT NULL REFERENCES absence_types(id),
    start_date DATE NOT NULL, -- from
    end_date   DATE NOT NULL, -- to
    total_days NUMERIC(5,2) NOT NULL,
    reason TEXT,
    status absence_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL DEFAULT now(), --  when is created
    created_by BIGINT NOT NULL REFERENCES employees(id), -- who is creator
    approved_at TIMESTAMP, -- when is approved?
    approved_by BIGINT NOT NULL REFERENCES employees(id), -- who is approver
    CONSTRAINT chk_dates CHECK (start_date <= end_date) -- start date cannot be greater
);

-- This returns the active policy configuration.
-- this application reads: grant policy, days_per_period, carry over rules
SELECT e.employee_id,
       e.absence_type_id,
       p.*
FROM employee_leave_policy e
JOIN leave_policies p ON p.id = e.policy_id
WHERE e.employee_id = 10
AND e.absence_type_id = 1
AND e.valid_from <= CURRENT_DATE
AND (e.valid_to IS NULL OR e.valid_to >= CURRENT_DATE);

SELECT p.*
FROM employee_leave_policy e
JOIN leave_policies p ON p.id = e.policy_id
WHERE e.employee_id = 10
AND e.absence_type_id = 1
AND e.valid_from <= DATE '2025-01-01'
AND (e.valid_to IS NULL OR e.valid_to >= DATE '2026-01-01');

-- LIST OF EMPLOYEES WITH THEIR ACTIVE VACATION POLICY
SELECT emp.id,
       emp.first_name,
       emp.last_name,
       p.name AS policy,
       p.days_per_period
FROM employee_leave_policy e
JOIN employees emp ON emp.id = e.employee_id
JOIN leave_policies p ON p.id = e.policy_id
WHERE e.absence_type_id = 1
AND e.valid_to IS NULL;

--- Employee - how request absence
--- Absence Type - type of absence is requested
--- Absence Request - API call, request from employee
--- POST and PUT for request and change status of absence request
--- Absence Balance - how much day are available for every
--- Absence History - audit log for absence of employee


-- every employee has position in firm that represent job which he works and at which level is employee job
create type POSITION_LEVEL as ENUM ('ROOKIE','JUNIOR','MEDIOR','SENIOR', 'LEAD')

-- Department → Positions → EmployeePosition → Employee
-- Each department has exactly one head.
create table if not exists departments(
  id bigint primary key not null,
  name varchar(30) not null,
  description text,
  STATUS status default 'ACTIVE',
  head_employee_id BIGINT UNIQUE REFERENCES employees(id)
);


CREATE TABLE positions (
    id BIGSERIAL PRIMARY KEY,
    departments_id BIGINT not null references departments(id),
    title VARCHAR(100) NOT NULL,         -- "Senior Backend Engineer"
    level POSITION_LEVEL NOT null default 'ROOKIE',          -- Junior/Mid/Senior/Lead
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
);

CREATE TABLE department_approvers (
    id BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL REFERENCES departments(id),
    employee_id BIGINT NOT NULL REFERENCES employees(id),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    valid_from DATE NOT NULL DEFAULT CURRENT_DATE,
    valid_to DATE
);

-- Department + Position are properties of the role at a specific time.
CREATE TABLE employees(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    phone_number VARCHAR(20),
    email_address VARCHAR(50),
    private_email VARCHAR(50),
    street VARCHAR(100),
    hire_date DATE NOT NULL,
    position_id BIGINT NOT NULL REFERENCES positions(id)
    created_at TIMESTAMP DEFAULT now()
);


-- OVO dodaj
CREATE TABLE employee_employment_history (
    id BIGSERIAL PRIMARY KEY,

    employee_id BIGINT NOT NULL REFERENCES employees(id),

    position_id BIGINT NOT NULL REFERENCES positions(id),

    start_date DATE NOT NULL,
    end_date DATE, -- NULL = currently active

    salary NUMERIC(12,2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'RSD',

    employment_type VARCHAR(20) NOT NULL,
    -- FULL_TIME / PART_TIME / CONTRACTOR / INTERN

    created_at TIMESTAMP NOT NULL DEFAULT now(),
    created_by BIGINT NOT NULL REFERENCES users(id),

    CONSTRAINT chk_dates CHECK (end_date IS NULL OR end_date >= start_date)
);
-- OVO POGLEDAJ NESTO NE STIMA!!!! - ovo ti realno ne treba
-- employee with position 'LEAD' is head of departmant
CREATE TABLE employee_positions (
    employee_position_id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    position_id BIGINT NOT NULL REFERENCES positions(id),
    salary NUMERIC(12,2) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT TRUE -- if is current head of departmant
);
-- for each request of employee - LEADs  or second in charge are main that can approve
CREATE TABLE department_absence_approvers (
    department_absence_approver_id BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL
        REFERENCES departments(id) ON DELETE CASCADE,
    employee_position_id BIGINT NOT NULL
        REFERENCES employee_positions(employee_position_id) ON DELETE CASCADE,

    approval_level INT NOT NULL DEFAULT 1,
        -- 1 = first approver
        -- 2 = HR / second level
        -- supports multi-step workflows later

    can_approve_paid BOOLEAN NOT NULL DEFAULT true,
    can_approve_unpaid BOOLEAN NOT NULL DEFAULT true,

    start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE,

    is_active BOOLEAN NOT NULL DEFAULT true
);



--- WORKFLOW CHECK
-- Create login
INSERT INTO users(email,password)
VALUES ('emilija@company.com','hash');

SELECT * FROM users;

-- Create employee
-- good, two employees with same user_id cannot exist - CHECKED
INSERT INTO employees
(user_id, first_name, last_name, hire_date, position_id)
VALUES (2,'Emilija','Ristic','2024-03-10',1);

SELECT * FROM employees e ;

--Assign policies to employee
-- Vacation policy assignment
INSERT INTO employee_leave_policy
(employee_id, absence_type_id, policy_id, valid_from)
VALUES (3, 1, 1, '2024-03-10');

-- Sick unlimited policy assignment
INSERT INTO employee_leave_policy
(employee_id, absence_type_id, policy_id, valid_from)
VALUES (3, 3, 2, '2024-03-10');

SELECT * FROM employee_leave_policy WHERE employee_id = 3;

-- Calculate prorated days for 2024
INSERT INTO leave_balance
(employee_id, absence_type_id, entry_type_leave, days, accrual_year, expires_at)
VALUES
(3,1,'ACCURAL',16.67,2024,'2025-06-30'); -- fix 16.67/??

SELECT * FROM leave_balance WHERE employee_id = 3;

-- january 1 - reset + carry over
INSERT INTO leave_balance
(employee_id, absence_type_id, entry_type_leave, days, accrual_year, expires_at)
VALUES
(1,1,'ACCURAL',20,2025,'2026-06-30');

-- check available days balance UI call
SELECT COALESCE(SUM(days),0) AS available_days
FROM leave_balance
WHERE employee_id = 1
AND absence_type_id = 1
AND (expires_at IS NULL OR expires_at >= CURRENT_DATE);

INSERT INTO absence_requests
(employee_id, absence_type_id, start_date, end_date, total_days, reason, created_by)
VALUES
(3,1,'2025-07-01','2025-07-05',5,'Summer vacation',1);
-- approved by can be nulL!
SELECT * FROM absence_requests WHERE employee_id = 3;

-- approved
UPDATE absence_requests
SET status='APPROVED',
    approved_at = now(),
    approved_by = 1
WHERE id = 1;

INSERT INTO leave_balance
(employee_id, absence_type_id, entry_type_leave, days, accrual_year, reference_id)
VALUES
(3,1,'CONSUMED',-5,2025,1);

SELECT * FROM leave_balance WHERE employee_id =3;

SELECT SUM(days)
FROM leave_balance
WHERE employee_id = 3
AND absence_type_id = 1
AND (expires_at IS NULL OR expires_at >= CURRENT_DATE);

-- Cancellation scenario (returns days)
UPDATE absence_requests
SET status='CANCELLED'
WHERE id=2;

INSERT INTO leave_balance
(employee_id, absence_type_id, entry_type_leave, days, accrual_year, reference_id)
VALUES
(3,1,'CANCELED',5,2025,2);

--STEP 10 — Carry-over job (Jan 1 2026)
SELECT SUM(days) FROM leave_balance
WHERE employee_id=3
AND absence_type_id=1
AND accrual_year=2024;

INSERT INTO leave_balance
(employee_id, absence_type_id, entry_type_leave, days, accrual_year)
VALUES
(3,1,'EXPIRATION',-10,2025);

-- current balance
SELECT SUM(days)
FROM leave_balance
WHERE employee_id=3
AND absence_type_id=1
AND (expires_at IS NULL OR expires_at >= CURRENT_DATE);

-- total used vacation
SELECT ABS(SUM(days))
FROM leave_balance
WHERE employee_id=1
AND absence_type_id=1
AND entry_type_leave='CONSUMED';
--Audit trail for one absence request
SELECT *
FROM leave_balance
WHERE reference_id = 1;
