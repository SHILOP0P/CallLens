package prompt_profile

import (
	"calllens/monolit/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }
func (r *Repository) Industries(ctx context.Context, perspective string) ([]models.PromptIndustry, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT industry_key,perspective,title,sort_order FROM prompt_industries WHERE ($1='' OR perspective=$1) ORDER BY perspective,sort_order,title`, perspective)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []models.PromptIndustry{}
	for rows.Next() {
		var x models.PromptIndustry
		e = rows.Scan(&x.Key, &x.Perspective, &x.Title, &x.SortOrder)
		if e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Topics(ctx context.Context, industry, q string) ([]models.PromptTopic, error) {
	q = strings.TrimSpace(strings.ToLower(q))
	rows, e := r.db.QueryContext(ctx, `SELECT topic_key,industry_key,title,prompt_module,sort_order FROM prompt_topics WHERE industry_key=$1 AND is_active AND ($2='' OR lower(title) LIKE '%' || $2 || '%') ORDER BY sort_order,title`, industry, q)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []models.PromptTopic{}
	for rows.Next() {
		var x models.PromptTopic
		e = rows.Scan(&x.Key, &x.IndustryKey, &x.Title, &x.PromptModule, &x.SortOrder)
		if e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Recommend(ctx context.Context, perspective, description string) ([]models.PromptTopic, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT DISTINCT ON (t.topic_key) t.topic_key,t.industry_key,t.title,t.prompt_module,t.sort_order FROM prompt_topic_aliases a JOIN prompt_topics t ON t.topic_key=a.topic_key JOIN prompt_industries i ON i.industry_key=t.industry_key WHERE ($1='' OR i.perspective=$1) AND lower($2) LIKE '%' || a.normalized_phrase || '%' AND NOT a.is_negative ORDER BY t.topic_key,a.weight DESC LIMIT 7`, perspective, strings.ToLower(description))
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []models.PromptTopic{}
	for rows.Next() {
		var x models.PromptTopic
		e = rows.Scan(&x.Key, &x.IndustryKey, &x.Title, &x.PromptModule, &x.SortOrder)
		if e != nil {
			return nil, e
		}
		x.Source = "recommended"
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *Repository) UserSettings(ctx context.Context, userID uuid.UUID) (models.PromptUserSettings, error) {
	settings := models.PromptUserSettings{UserID: userID, Industries: []models.PromptIndustry{}, Topics: []models.PromptTopic{}}
	_ = r.db.QueryRowContext(ctx, `SELECT description FROM prompt_user_settings WHERE user_uuid=$1`, userID).Scan(&settings.Description)
	rows, err := r.db.QueryContext(ctx, `SELECT i.industry_key,i.perspective,i.title,i.sort_order FROM prompt_user_industries u JOIN prompt_industries i ON i.industry_key=u.industry_key WHERE u.user_uuid=$1 ORDER BY u.sort_order,i.sort_order`, userID)
	if err != nil {
		return settings, err
	}
	defer rows.Close()
	for rows.Next() {
		var item models.PromptIndustry
		if err = rows.Scan(&item.Key, &item.Perspective, &item.Title, &item.SortOrder); err != nil {
			return settings, err
		}
		settings.Industries = append(settings.Industries, item)
	}
	if err = rows.Err(); err != nil {
		return settings, err
	}
	rows, err = r.db.QueryContext(ctx, `SELECT t.topic_key,t.industry_key,t.title,t.prompt_module,t.sort_order,u.source FROM prompt_user_topics u JOIN prompt_topics t ON t.topic_key=u.topic_key WHERE u.user_uuid=$1 ORDER BY u.sort_order,t.sort_order`, userID)
	if err != nil {
		return settings, err
	}
	defer rows.Close()
	for rows.Next() {
		var item models.PromptTopic
		if err = rows.Scan(&item.Key, &item.IndustryKey, &item.Title, &item.PromptModule, &item.SortOrder, &item.Source); err != nil {
			return settings, err
		}
		settings.Topics = append(settings.Topics, item)
	}
	return settings, rows.Err()
}
func (r *Repository) SaveUserSettings(ctx context.Context, settings models.PromptUserSettings) (models.PromptUserSettings, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return settings, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO prompt_user_settings(user_uuid,description) VALUES($1,$2) ON CONFLICT(user_uuid) DO UPDATE SET description=EXCLUDED.description,updated_at=now()`, settings.UserID, settings.Description); err != nil {
		return settings, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM prompt_user_industries WHERE user_uuid=$1`, settings.UserID); err != nil {
		return settings, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM prompt_user_topics WHERE user_uuid=$1`, settings.UserID); err != nil {
		return settings, err
	}
	for n, item := range settings.Industries {
		if _, err = tx.ExecContext(ctx, `INSERT INTO prompt_user_industries(user_uuid,industry_key,sort_order) VALUES($1,$2,$3)`, settings.UserID, item.Key, n); err != nil {
			return settings, err
		}
	}
	for n, item := range settings.Topics {
		if _, err = tx.ExecContext(ctx, `INSERT INTO prompt_user_topics(user_uuid,topic_key,source,sort_order) VALUES($1,$2,$3,$4)`, settings.UserID, item.Key, coalesce(item.Source, "manual"), n); err != nil {
			return settings, err
		}
	}
	if err = tx.Commit(); err != nil {
		return settings, err
	}
	return r.UserSettings(ctx, settings.UserID)
}
func (r *Repository) ListProfiles(ctx context.Context, owner uuid.UUID) ([]models.PromptProfile, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT profile_uuid,title,perspective,industry_key,description,is_default FROM prompt_profiles WHERE owner_user_uuid=$1 ORDER BY is_default DESC,updated_at DESC`, owner)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []models.PromptProfile{}
	for rows.Next() {
		var x models.PromptProfile
		x.OwnerUserID = owner
		if e = rows.Scan(&x.ID, &x.Title, &x.Perspective, &x.IndustryKey, &x.Description, &x.IsDefault); e != nil {
			return nil, e
		}
		x.Topics, _ = r.profileTopics(ctx, x.ID)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) profileTopics(ctx context.Context, id uuid.UUID) ([]models.PromptTopic, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT t.topic_key,t.industry_key,t.title,t.prompt_module,t.sort_order,p.source FROM prompt_profile_topics p JOIN prompt_topics t ON t.topic_key=p.topic_key WHERE p.profile_uuid=$1 ORDER BY p.sort_order,t.sort_order`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []models.PromptTopic{}
	for rows.Next() {
		var x models.PromptTopic
		if e = rows.Scan(&x.Key, &x.IndustryKey, &x.Title, &x.PromptModule, &x.SortOrder, &x.Source); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) SaveProfile(ctx context.Context, p models.PromptProfile) (models.PromptProfile, error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	tx, e := r.db.BeginTx(ctx, nil)
	if e != nil {
		return p, e
	}
	defer tx.Rollback()
	if p.IsDefault {
		if _, e = tx.ExecContext(ctx, `UPDATE prompt_profiles SET is_default=false WHERE owner_user_uuid=$1 AND perspective=$2`, p.OwnerUserID, p.Perspective); e != nil {
			return p, e
		}
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO prompt_profiles(profile_uuid,owner_user_uuid,title,perspective,industry_key,description,is_default) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(profile_uuid) DO UPDATE SET title=EXCLUDED.title,perspective=EXCLUDED.perspective,industry_key=EXCLUDED.industry_key,description=EXCLUDED.description,is_default=EXCLUDED.is_default,updated_at=now() WHERE prompt_profiles.owner_user_uuid=$2`, p.ID, p.OwnerUserID, p.Title, p.Perspective, p.IndustryKey, p.Description, p.IsDefault)
	if e != nil {
		return p, e
	}
	if _, e = tx.ExecContext(ctx, `DELETE FROM prompt_profile_topics WHERE profile_uuid=$1`, p.ID); e != nil {
		return p, e
	}
	for n, t := range p.Topics {
		_, e = tx.ExecContext(ctx, `INSERT INTO prompt_profile_topics(profile_uuid,topic_key,source,sort_order) VALUES($1,$2,$3,$4)`, p.ID, t.Key, coalesce(t.Source, "manual"), n)
		if e != nil {
			return p, e
		}
	}
	if e = tx.Commit(); e != nil {
		return p, e
	}
	p.Topics, _ = r.profileTopics(ctx, p.ID)
	return p, nil
}
func coalesce(a, b string) string {
	if a == "" {
		return b
	}
	return a
}
func (r *Repository) DeleteProfile(ctx context.Context, id, owner uuid.UUID) error {
	res, e := r.db.ExecContext(ctx, `DELETE FROM prompt_profiles WHERE profile_uuid=$1 AND owner_user_uuid=$2`, id, owner)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("profile not found")
	}
	return nil
}
func (r *Repository) PutCallContext(ctx context.Context, c models.CallPromptContext) error {
	raw, e := json.Marshal(c.TopicKeys)
	if e != nil {
		return e
	}
	_, e = r.db.ExecContext(ctx, `INSERT INTO call_prompt_contexts(call_uuid,profile_uuid,owner_user_uuid,topic_keys) VALUES($1,NULLIF($2,'00000000-0000-0000-0000-000000000000'),$3,$4::jsonb) ON CONFLICT(call_uuid) DO UPDATE SET profile_uuid=EXCLUDED.profile_uuid,owner_user_uuid=EXCLUDED.owner_user_uuid,topic_keys=EXCLUDED.topic_keys,updated_at=now()`, c.CallID, c.ProfileID, c.OwnerUserID, string(raw))
	return e
}
func (r *Repository) CallContext(ctx context.Context, callID, owner uuid.UUID) (models.CallPromptContext, error) {
	var c models.CallPromptContext
	var raw []byte
	c.CallID = callID
	c.OwnerUserID = owner
	e := r.db.QueryRowContext(ctx, `SELECT COALESCE(profile_uuid,'00000000-0000-0000-0000-000000000000'::uuid),topic_keys FROM call_prompt_contexts WHERE call_uuid=$1 AND owner_user_uuid=$2`, callID, owner).Scan(&c.ProfileID, &raw)
	if e != nil {
		return c, e
	}
	e = json.Unmarshal(raw, &c.TopicKeys)
	return c, e
}
func (r *Repository) Modules(ctx context.Context, callID uuid.UUID, userID uuid.UUID) ([]models.PromptTopic, error) {
	var snapshot []byte
	err := r.db.QueryRowContext(ctx, `SELECT topics_json FROM analysis_prompt_snapshots WHERE call_uuid=$1`, callID).Scan(&snapshot)
	if err == nil {
		var topics []models.PromptTopic
		if err = json.Unmarshal(snapshot, &topics); err != nil {
			return nil, err
		}
		return topics, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	return r.contextModules(ctx, callID, userID)
}

func (r *Repository) Snapshot(ctx context.Context, callID uuid.UUID, userID uuid.UUID) error {
	topics, err := r.contextModules(ctx, callID, userID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(topics)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO analysis_prompt_snapshots(call_uuid,topics_json) VALUES($1,$2::jsonb) ON CONFLICT(call_uuid) DO UPDATE SET topics_json=EXCLUDED.topics_json,created_at=now()`, callID, string(payload))
	return err
}

func (r *Repository) contextModules(ctx context.Context, callID uuid.UUID, userID uuid.UUID) ([]models.PromptTopic, error) {
	rows, e := r.db.QueryContext(ctx, `WITH has_context AS (SELECT EXISTS(SELECT 1 FROM call_prompt_contexts WHERE call_uuid=$1) AS value), selected AS (SELECT jsonb_array_elements_text(topic_keys) AS topic_key FROM call_prompt_contexts WHERE call_uuid=$1 UNION SELECT topic_key FROM prompt_user_topics WHERE user_uuid=$2 AND NOT (SELECT value FROM has_context)), industry_modules AS (SELECT 'industry-' || i.industry_key,i.industry_key,i.title,i.base_prompt,i.sort_order FROM prompt_user_industries u JOIN prompt_industries i ON i.industry_key=u.industry_key WHERE u.user_uuid=$2 AND NOT (SELECT value FROM has_context)) SELECT DISTINCT ON (key) key,industry_key,title,prompt_module,sort_order FROM (SELECT t.topic_key AS key,t.industry_key,t.title,t.prompt_module,t.sort_order FROM selected s JOIN prompt_topics t ON t.topic_key=s.topic_key WHERE t.is_active UNION ALL SELECT * FROM industry_modules) modules ORDER BY key,sort_order`, callID, userID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []models.PromptTopic{}
	for rows.Next() {
		var x models.PromptTopic
		e = rows.Scan(&x.Key, &x.IndustryKey, &x.Title, &x.PromptModule, &x.SortOrder)
		if e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
