DROP TRIGGER notify_event ON organizations;
---- create above / drop below ----
CREATE OR REPLACE TRIGGER notify_event AFTER INSERT OR DELETE OR UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION build_and_send_event();
