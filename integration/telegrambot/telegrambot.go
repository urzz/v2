// Copyright 2021 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package telegrambot // import "miniflux.app/integration/telegrambot"

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"miniflux.app/model"
)

// PushEntry pushes entry to telegram chat using integration settings provided
func PushEntry(entry *model.Entry, botToken, chatID, proxyUrl string) error {
	client := &http.Client{}
	trimProxyUrl := strings.TrimSpace(proxyUrl)
	if len(trimProxyUrl) > 0 {
		proxy, err := url.Parse(trimProxyUrl)
		if err != nil {
			return fmt.Errorf("telegrambot: bot creation failed, proxyUrl invalid: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	bot, err := tgbotapi.NewBotAPIWithClient(botToken, client)
	if err != nil {
		return fmt.Errorf("telegrambot: bot creation failed: %w", err)
	}

	tpl, err := template.New("message").Parse("{{ .Title }}\n<a href=\"{{ .URL }}\">{{ .URL }}</a>")
	if err != nil {
		return fmt.Errorf("telegrambot: template parsing failed: %w", err)
	}

	var result bytes.Buffer
	if err := tpl.Execute(&result, entry); err != nil {
		return fmt.Errorf("telegrambot: template execution failed: %w", err)
	}

	chatIDInt, _ := strconv.ParseInt(chatID, 10, 64)
	msg := tgbotapi.NewMessage(chatIDInt, result.String())
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = false
	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("telegrambot: sending message failed: %w", err)
	}

	return nil
}
