// Package model provides ...
package model

import (
	"encoding/json"
	"testing"

	"github.com/urfave/cli/v2"
)

var ddlSQLs = []string{
	`CREATE TABLE "user" (
	    id         SERIAL                              NOT NULL,
	    name       TEXT      DEFAULT 1                 NOT NULL,
	    age        INTEGER                             NOT NULL,
	    email      TEXT                                NOT NULL,
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	    UNIQUE (email, age),
		PRIMARY KEY (id)
	);`,
	`CREATE TABLE IF NOT EXISTS "fido_credential" (
	    id SERIAL NOT NULL,
	    tenant_id INTEGER NOT NULL,

	    credential_id TEXT NOT NULL,
	    user_id TEXT NOT NULL,
	    public_key bytea NOT NULL,
	    authenticator_id INTEGER NOT NULL,
	    sign_count INTEGER NOT NULL DEFAULT 0,
	    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	    PRIMARY KEY (id)
	);
	COMMENT ON TABLE fido_credential IS "凭证表";
	COMMENT ON COLUMN fido_credential.id IS "自增ID";
	COMMENT ON COLUMN fido_credential.tenant_id IS "租户ID";
	COMMENT ON COLUMN fido_credential.credential_id IS "凭证ID";
	COMMENT ON COLUMN fido_credential.user_id IS "用户ID";
	COMMENT ON COLUMN fido_credential.public_key IS "公钥";
	COMMENT ON COLUMN fido_credential.authenticator_id IS "认证器ID";
	COMMENT ON COLUMN fido_credential.sign_count IS "签名次数";
	COMMENT ON COLUMN fido_credential.updated_at IS "更新时间";
	COMMENT ON COLUMN fido_credential.created_at IS "创建时间";
	`,
	`CREATE TABLE IF NOT EXISTS "source"
(
    id         SERIAL NOT NULL,
    platform   INTEGER,
    version_id INTEGER,
    channel    INTEGER,
    url        VARCHAR(1024),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delete_at  TIMESTAMP,
    PRIMARY KEY (id)
);

COMMENT ON TABLE source IS '包';
COMMENT ON COLUMN source.id IS 'id';
COMMENT ON COLUMN source.platform IS '平台，1安卓，2IOS';
COMMENT ON COLUMN source.version_id IS '版本id';
COMMENT ON COLUMN source.channel IS '渠道，0官方，1.google，2.huawei，3.xiaomi，4.oppo，5.vivo，6.魅族，7.应用宝，8.360手机助手,9.豌豆荚,10.酷安';
COMMENT ON COLUMN source.url IS '下载地址';
COMMENT ON COLUMN source.created_at IS '创建时间';
COMMENT ON COLUMN source.updated_at IS '修改时间';
COMMENT ON COLUMN source.delete_at IS '删除时间';`,
}

func TestDDLParse(t *testing.T) {
	for _, v := range ddlSQLs {
		params, err := ddlAnalyzer([]byte(v))
		if err != nil {
			t.Fatal(err)
		}
		data, _ := json.MarshalIndent(params, "", "  ")
		t.Log(string(data))
	}
}

func TestCommandModel(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		ModelCommand,
	}

	err := app.Run([]string{"zero", "model", "--src", "testdata"})
	if err != nil {
		t.Fatal(err)
	}
}
