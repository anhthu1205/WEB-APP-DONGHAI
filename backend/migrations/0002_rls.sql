DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['schema_migrations','users','company_info','products','product_images',
                           'promotions','promotion_items','customers','orders','order_items'] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
      EXECUTE format('REVOKE ALL ON TABLE %I FROM anon, authenticated', t);
    END IF;
  END LOOP;
END $$;
INSERT INTO schema_migrations (version) VALUES ('0002_rls') ON CONFLICT DO NOTHING;
