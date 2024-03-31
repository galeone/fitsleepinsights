-- alter table statements used to create relationships among tables
-- that are not possible during the creation (for dependency issues)

-- nullable
ALTER TABLE heart_rate_zones ADD COLUMN IF NOT EXISTS heart_rate_activity_id bigint references heart_rate_activities(id);
ALTER TABLE oauth2_authorized ADD COLUMN IF NOT EXISTS dumping BOOLEAN not null default true;

-- series indexes
CREATE INDEX IF NOT EXISTS activity_calories_series_idx ON activity_calories_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS bmi_series_idx ON bmi_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS body_fat_series_idx ON body_fat_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS body_weight_series_idx ON body_weight_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS calories_bmr_series_idx ON calories_bmr_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS calories_series_idx ON calories_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS distance_series_idx ON distance_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS elevation_series_idx ON elevation_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS floors_series_idx ON floors_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS heart_rate_variability_time_series_idx ON heart_rate_variability_time_series (id, user_id, "date");
CREATE INDEX IF NOT EXISTS minutes_fairly_active_series_idx ON minutes_fairly_active_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS minutes_lightly_active_series_idx ON minutes_lightly_active_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS minutes_sedentary_series_idx ON minutes_sedentary_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS minutes_very_active_series_idx ON minutes_very_active_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS steps_series_idx ON steps_series (id, "date", user_id);
CREATE INDEX IF NOT EXISTS heart_rate_activities_idx ON heart_rate_activities (id, user_id, "date");
CREATE INDEX IF NOT EXISTS cardio_fitness_score_idx ON cardio_fitness_score (id, user_id, "date");
CREATE INDEX IF NOT EXISTS core_temperatures_idx ON core_temperatures (id, user_id, "date");
CREATE INDEX IF NOT EXISTS oxygen_saturation_idx ON oxygen_saturation (id, user_id, "date");
CREATE INDEX IF NOT EXISTS skin_temperatures_idx ON skin_temperatures (id, user_id, "date");

-- activity indexes
CREATE INDEX IF NOT EXISTS activity_logs_idx ON activity_logs (log_id, user_id, date(start_time));
CREATE INDEX IF NOT EXISTS heart_rate_zones_idx ON heart_rate_zones (id, activity_log_id, heart_rate_activity_id, "type");
CREATE INDEX IF NOT EXISTS minutes_in_heart_rate_zone_idx ON minutes_in_heart_rate_zone (id, active_zone_minutes_id);

-- sleep indexes
CREATE INDEX IF NOT EXISTS sleep_data_idx ON sleep_data (sleep_log_id);
CREATE INDEX IF NOT EXISTS sleep_logs_idx ON sleep_logs (date_of_sleep, user_id);
CREATE INDEX IF NOT EXISTS sleep_stage_details_idx ON sleep_stage_details (sleep_log_id);