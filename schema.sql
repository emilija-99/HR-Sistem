--
-- PostgreSQL database dump
--

\restrict nyJlmRNfya1rrNcsSLovP3cDWlsm4yEu1F7lsRhrEDFDH8ZWPaD0PKDvYOPTATs

-- Dumped from database version 15.17 (Debian 15.17-1.pgdg13+1)
-- Dumped by pg_dump version 18.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: absence_status; Type: TYPE; Schema: public; Owner: hr_user
--

CREATE TYPE public.absence_status AS ENUM (
    'PENDING',
    'APPROVED',
    'REJECTED',
    'CANCELLED'
);


ALTER TYPE public.absence_status OWNER TO hr_user;

--
-- Name: action_audit_log; Type: TYPE; Schema: public; Owner: hr_user
--

CREATE TYPE public.action_audit_log AS ENUM (
    'INSERT',
    'UPDATE',
    'DELETE',
    'LINK',
    'UNLINK',
    'APPROVE',
    'REJECT'
);


ALTER TYPE public.action_audit_log OWNER TO hr_user;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: absence_types; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.absence_types (
    absence_type_id integer NOT NULL,
    type_name character varying(40) NOT NULL,
    code character varying(40) NOT NULL,
    is_paid boolean NOT NULL
);


ALTER TABLE public.absence_types OWNER TO hr_user;

--
-- Name: absence_types_history; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.absence_types_history (
    history_id bigint NOT NULL,
    absence_type_id bigint NOT NULL,
    absence_status public.absence_status NOT NULL,
    valid_from timestamp without time zone NOT NULL,
    valid_to timestamp without time zone,
    changed_at timestamp without time zone DEFAULT now() NOT NULL,
    approved_by bigint,
    approved_at timestamp without time zone
);


ALTER TABLE public.absence_types_history OWNER TO hr_user;

--
-- Name: absence_types_history_history_id_seq; Type: SEQUENCE; Schema: public; Owner: hr_user
--

CREATE SEQUENCE public.absence_types_history_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.absence_types_history_history_id_seq OWNER TO hr_user;

--
-- Name: absence_types_history_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hr_user
--

ALTER SEQUENCE public.absence_types_history_history_id_seq OWNED BY public.absence_types_history.history_id;


--
-- Name: audit_log; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.audit_log (
    audit_id bigint NOT NULL,
    entity_name character varying(50) NOT NULL,
    entity_id bigint NOT NULL,
    action public.action_audit_log NOT NULL,
    old_data jsonb,
    new_data jsonb,
    changed_by bigint,
    changed_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.audit_log OWNER TO hr_user;

--
-- Name: audit_log_audit_id_seq; Type: SEQUENCE; Schema: public; Owner: hr_user
--

CREATE SEQUENCE public.audit_log_audit_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.audit_log_audit_id_seq OWNER TO hr_user;

--
-- Name: audit_log_audit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hr_user
--

ALTER SEQUENCE public.audit_log_audit_id_seq OWNED BY public.audit_log.audit_id;


--
-- Name: bank_accounts; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.bank_accounts (
    bank_account_id bigint NOT NULL,
    employee_id bigint NOT NULL,
    bank_name character varying(50) NOT NULL,
    account_number character varying(50) NOT NULL,
    is_primary boolean DEFAULT false NOT NULL
);


ALTER TABLE public.bank_accounts OWNER TO hr_user;

--
-- Name: bank_accounts_bank_account_id_seq; Type: SEQUENCE; Schema: public; Owner: hr_user
--

CREATE SEQUENCE public.bank_accounts_bank_account_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.bank_accounts_bank_account_id_seq OWNER TO hr_user;

--
-- Name: bank_accounts_bank_account_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hr_user
--

ALTER SEQUENCE public.bank_accounts_bank_account_id_seq OWNED BY public.bank_accounts.bank_account_id;


--
-- Name: countries; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.countries (
    country_id integer NOT NULL,
    country_name character varying(40) NOT NULL
);


ALTER TABLE public.countries OWNER TO hr_user;

--
-- Name: departmants; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.departmants (
    department_id bigint NOT NULL,
    department_name character varying(30) NOT NULL,
    manager_id integer
);


ALTER TABLE public.departmants OWNER TO hr_user;

--
-- Name: education_levels; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.education_levels (
    education_level_id integer NOT NULL,
    name character varying(100) NOT NULL
);


ALTER TABLE public.education_levels OWNER TO hr_user;

--
-- Name: education_levels_education_level_id_seq; Type: SEQUENCE; Schema: public; Owner: hr_user
--

CREATE SEQUENCE public.education_levels_education_level_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.education_levels_education_level_id_seq OWNER TO hr_user;

--
-- Name: education_levels_education_level_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hr_user
--

ALTER SEQUENCE public.education_levels_education_level_id_seq OWNED BY public.education_levels.education_level_id;


--
-- Name: employee_children; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.employee_children (
    child_id bigint NOT NULL,
    employee_id bigint NOT NULL,
    first_name character varying(50),
    birth_date date NOT NULL
);


ALTER TABLE public.employee_children OWNER TO hr_user;

--
-- Name: employee_children_child_id_seq; Type: SEQUENCE; Schema: public; Owner: hr_user
--

CREATE SEQUENCE public.employee_children_child_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.employee_children_child_id_seq OWNER TO hr_user;

--
-- Name: employee_children_child_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hr_user
--

ALTER SEQUENCE public.employee_children_child_id_seq OWNED BY public.employee_children.child_id;


--
-- Name: employee_education; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.employee_education (
    education_id bigint NOT NULL,
    employee_id bigint NOT NULL,
    education_level_id integer,
    institution_name character varying(200),
    field_of_study character varying(150),
    start_year integer,
    end_year integer
);


ALTER TABLE public.employee_education OWNER TO hr_user;

--
-- Name: employee_education_education_id_seq; Type: SEQUENCE; Schema: public; Owner: hr_user
--

CREATE SEQUENCE public.employee_education_education_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.employee_education_education_id_seq OWNER TO hr_user;

--
-- Name: employee_education_education_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hr_user
--

ALTER SEQUENCE public.employee_education_education_id_seq OWNED BY public.employee_education.education_id;


--
-- Name: employee_social_data; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.employee_social_data (
    social_id bigint NOT NULL,
    employee_id bigint NOT NULL,
    jmbg character varying(13) NOT NULL,
    tax_id character varying(30),
    health_insurance_number character varying(30)
);


ALTER TABLE public.employee_social_data OWNER TO hr_user;

--
-- Name: employee_social_data_social_id_seq; Type: SEQUENCE; Schema: public; Owner: hr_user
--

CREATE SEQUENCE public.employee_social_data_social_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.employee_social_data_social_id_seq OWNER TO hr_user;

--
-- Name: employee_social_data_social_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hr_user
--

ALTER SEQUENCE public.employee_social_data_social_id_seq OWNED BY public.employee_social_data.social_id;


--
-- Name: employees; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.employees (
    employee_id bigint NOT NULL,
    user_id bigint NOT NULL,
    first_name character varying(20) NOT NULL,
    last_name character varying(20) NOT NULL,
    phone_number character varying(20),
    hire_date date NOT NULL,
    job_id bigint NOT NULL,
    salary double precision NOT NULL,
    department_id bigint NOT NULL,
    manager_id integer
);


ALTER TABLE public.employees OWNER TO hr_user;

--
-- Name: job_history; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.job_history (
    employee_id bigint NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    job_id bigint NOT NULL,
    department_id bigint
);


ALTER TABLE public.job_history OWNER TO hr_user;

--
-- Name: jobs; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.jobs (
    job_id bigint NOT NULL,
    job_title character varying(35) NOT NULL,
    min_salary double precision,
    max_salary double precision
);


ALTER TABLE public.jobs OWNER TO hr_user;

--
-- Name: locations; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.locations (
    location_id bigint NOT NULL,
    street_address character varying(40),
    postal_code character varying(12),
    city character varying(30) NOT NULL,
    country_id integer
);


ALTER TABLE public.locations OWNER TO hr_user;

--
-- Name: permissions; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.permissions (
    id integer NOT NULL,
    name character varying(100) NOT NULL
);


ALTER TABLE public.permissions OWNER TO hr_user;

--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: hr_user
--

CREATE SEQUENCE public.permissions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.permissions_id_seq OWNER TO hr_user;

--
-- Name: permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hr_user
--

ALTER SEQUENCE public.permissions_id_seq OWNED BY public.permissions.id;


--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.refresh_tokens (
    id uuid NOT NULL,
    user_id bigint NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    revoked boolean DEFAULT false NOT NULL,
    user_agent text,
    ip_address text
);


ALTER TABLE public.refresh_tokens OWNER TO hr_user;

--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.role_permissions (
    role_id integer NOT NULL,
    permission_id integer NOT NULL
);


ALTER TABLE public.role_permissions OWNER TO hr_user;

--
-- Name: roles; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.roles (
    id integer NOT NULL,
    name character varying(50) NOT NULL
);


ALTER TABLE public.roles OWNER TO hr_user;

--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: hr_user
--

CREATE SEQUENCE public.roles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.roles_id_seq OWNER TO hr_user;

--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hr_user
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO hr_user;

--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.user_roles (
    user_id bigint NOT NULL,
    role_id integer NOT NULL
);


ALTER TABLE public.user_roles OWNER TO hr_user;

--
-- Name: users; Type: TABLE; Schema: public; Owner: hr_user
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    first_name character varying(20) NOT NULL,
    last_name character varying(20) NOT NULL,
    email character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.users OWNER TO hr_user;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: hr_user
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO hr_user;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hr_user
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: absence_types_history history_id; Type: DEFAULT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.absence_types_history ALTER COLUMN history_id SET DEFAULT nextval('public.absence_types_history_history_id_seq'::regclass);


--
-- Name: audit_log audit_id; Type: DEFAULT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.audit_log ALTER COLUMN audit_id SET DEFAULT nextval('public.audit_log_audit_id_seq'::regclass);


--
-- Name: bank_accounts bank_account_id; Type: DEFAULT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.bank_accounts ALTER COLUMN bank_account_id SET DEFAULT nextval('public.bank_accounts_bank_account_id_seq'::regclass);


--
-- Name: education_levels education_level_id; Type: DEFAULT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.education_levels ALTER COLUMN education_level_id SET DEFAULT nextval('public.education_levels_education_level_id_seq'::regclass);


--
-- Name: employee_children child_id; Type: DEFAULT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_children ALTER COLUMN child_id SET DEFAULT nextval('public.employee_children_child_id_seq'::regclass);


--
-- Name: employee_education education_id; Type: DEFAULT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_education ALTER COLUMN education_id SET DEFAULT nextval('public.employee_education_education_id_seq'::regclass);


--
-- Name: employee_social_data social_id; Type: DEFAULT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_social_data ALTER COLUMN social_id SET DEFAULT nextval('public.employee_social_data_social_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: absence_types absence_types_code_key; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.absence_types
    ADD CONSTRAINT absence_types_code_key UNIQUE (code);


--
-- Name: absence_types_history absence_types_history_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.absence_types_history
    ADD CONSTRAINT absence_types_history_pkey PRIMARY KEY (history_id);


--
-- Name: absence_types absence_types_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.absence_types
    ADD CONSTRAINT absence_types_pkey PRIMARY KEY (absence_type_id);


--
-- Name: absence_types absence_types_type_name_key; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.absence_types
    ADD CONSTRAINT absence_types_type_name_key UNIQUE (type_name);


--
-- Name: audit_log audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.audit_log
    ADD CONSTRAINT audit_log_pkey PRIMARY KEY (audit_id);


--
-- Name: bank_accounts bank_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.bank_accounts
    ADD CONSTRAINT bank_accounts_pkey PRIMARY KEY (bank_account_id);


--
-- Name: countries countries_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.countries
    ADD CONSTRAINT countries_pkey PRIMARY KEY (country_id);


--
-- Name: departmants departmants_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.departmants
    ADD CONSTRAINT departmants_pkey PRIMARY KEY (department_id);


--
-- Name: education_levels education_levels_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.education_levels
    ADD CONSTRAINT education_levels_pkey PRIMARY KEY (education_level_id);


--
-- Name: employee_children employee_children_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_children
    ADD CONSTRAINT employee_children_pkey PRIMARY KEY (child_id);


--
-- Name: employee_education employee_education_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_education
    ADD CONSTRAINT employee_education_pkey PRIMARY KEY (education_id);


--
-- Name: employee_social_data employee_social_data_employee_id_key; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_social_data
    ADD CONSTRAINT employee_social_data_employee_id_key UNIQUE (employee_id);


--
-- Name: employee_social_data employee_social_data_jmbg_key; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_social_data
    ADD CONSTRAINT employee_social_data_jmbg_key UNIQUE (jmbg);


--
-- Name: employee_social_data employee_social_data_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_social_data
    ADD CONSTRAINT employee_social_data_pkey PRIMARY KEY (social_id);


--
-- Name: employees employees_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_pkey PRIMARY KEY (employee_id);


--
-- Name: job_history job_history_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.job_history
    ADD CONSTRAINT job_history_pkey PRIMARY KEY (employee_id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (job_id);


--
-- Name: locations locations_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_pkey PRIMARY KEY (location_id);


--
-- Name: permissions permissions_name_key; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_name_key UNIQUE (name);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_name_key; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_name_key UNIQUE (name);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_refresh_user; Type: INDEX; Schema: public; Owner: hr_user
--

CREATE INDEX idx_refresh_user ON public.refresh_tokens USING btree (user_id);


--
-- Name: absence_types_history absence_types_history_absence_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.absence_types_history
    ADD CONSTRAINT absence_types_history_absence_type_id_fkey FOREIGN KEY (absence_type_id) REFERENCES public.absence_types(absence_type_id);


--
-- Name: absence_types_history absence_types_history_approved_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.absence_types_history
    ADD CONSTRAINT absence_types_history_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES public.users(id);


--
-- Name: audit_log audit_log_changed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.audit_log
    ADD CONSTRAINT audit_log_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES public.users(id);


--
-- Name: bank_accounts bank_accounts_employee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.bank_accounts
    ADD CONSTRAINT bank_accounts_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES public.employees(employee_id) ON DELETE CASCADE;


--
-- Name: employee_children employee_children_employee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_children
    ADD CONSTRAINT employee_children_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES public.employees(employee_id) ON DELETE CASCADE;


--
-- Name: employee_education employee_education_education_level_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_education
    ADD CONSTRAINT employee_education_education_level_id_fkey FOREIGN KEY (education_level_id) REFERENCES public.education_levels(education_level_id);


--
-- Name: employee_education employee_education_employee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_education
    ADD CONSTRAINT employee_education_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES public.employees(employee_id) ON DELETE CASCADE;


--
-- Name: employee_social_data employee_social_data_employee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employee_social_data
    ADD CONSTRAINT employee_social_data_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES public.employees(employee_id) ON DELETE CASCADE;


--
-- Name: employees employees_department_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.departmants(department_id);


--
-- Name: employees employees_job_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(job_id);


--
-- Name: employees employees_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: job_history job_history_department_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.job_history
    ADD CONSTRAINT job_history_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.departmants(department_id);


--
-- Name: locations locations_country_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_country_id_fkey FOREIGN KEY (country_id) REFERENCES public.countries(country_id);


--
-- Name: refresh_tokens refresh_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: hr_user
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict nyJlmRNfya1rrNcsSLovP3cDWlsm4yEu1F7lsRhrEDFDH8ZWPaD0PKDvYOPTATs

