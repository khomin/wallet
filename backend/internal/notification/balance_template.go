package notification

import (
	"bytes"
	"fmt"
	"html/template"
)

type BalanceTemplateData struct {
	UserName         string
	WalletName       string
	CoinSymbol       string // e.g. "BTC", "ETH", "USDT"
	FormattedBalance string // Current balance string, e.g. "1.4500"
	FormattedDelta   string // e.g. "+0.2500" or "-0.1000"
	IsDeposit        bool   // true for deposit (+), false for withdrawal (-)
}

var balanceEmailTmpl = template.Must(template.New("balanceEmail").Parse(`
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="utf-8">
    </head>
    <body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #0f172a; color: #f8fafc; margin: 0; padding: 24px;">
        <div style="max-width: 560px; margin: 0 auto; background-color: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 32px;">
            
            {{if .IsDeposit}}
            <div style="display: inline-block; background-color: rgba(34, 197, 94, 0.1); color: #4ade80; border: 1px solid rgba(74, 222, 128, 0.2); font-size: 12px; font-weight: 600; padding: 4px 12px; border-radius: 9999px; text-transform: uppercase; margin-bottom: 16px;">
                Deposit Detected
            </div>
            {{else}}
            <div style="display: inline-block; background-color: rgba(239, 68, 68, 0.1); color: #f87171; border: 1px solid rgba(248, 113, 113, 0.2); font-size: 12px; font-weight: 600; padding: 4px 12px; border-radius: 9999px; text-transform: uppercase; margin-bottom: 16px;">
                Withdrawal Detected
            </div>
            {{end}}
            
            <h1 style="font-size: 20px; font-weight: 600; margin: 0 0 12px 0; color: #ffffff;">
                Balance Update for {{.WalletName}}
            </h1>
            
            <p style="font-size: 14px; line-height: 1.6; color: #94a3b8; margin: 0 0 24px 0;">
                Hi {{.UserName}}, a balance change was detected on your wallet <strong>{{.WalletName}}</strong>.
            </p>
            
            <!-- Bulletproof Table with inline styles and explicit cell widths -->
            <table width="100%" cellpadding="0" cellspacing="0" border="0" style="width: 100%; min-width: 100%; background-color: #0f172a; border: 1px solid #334155; border-radius: 8px; margin-bottom: 24px; border-collapse: collapse;">
                <tr>
                    <td align="left" valign="middle" style="padding: 16px 20px; font-size: 14px; font-weight: 600; color: #94a3b8; text-align: left; width: 50%; border-bottom: 1px solid #1e293b;">
                        Change Amount
                    </td>
                    <td align="right" valign="middle" style="padding: 16px 20px; font-size: 16px; font-weight: 700; text-align: right; width: 50%; border-bottom: 1px solid #1e293b; {{if .IsDeposit}}color: #4ade80;{{else}}color: #f87171;{{end}}">
                        {{.FormattedDelta}} {{.CoinSymbol}}
                    </td>
                </tr>
                <tr>
                    <td align="left" valign="middle" style="padding: 16px 20px; font-size: 14px; font-weight: 600; color: #94a3b8; text-align: left; width: 50%;">
                        Current Balance
                    </td>
                    <td align="right" valign="middle" style="padding: 16px 20px; font-size: 18px; font-weight: 700; color: #38bdf8; text-align: right; width: 50%;">
                        {{.FormattedBalance}} {{.CoinSymbol}}
                    </td>
                </tr>
            </table>
            
            <p style="font-size: 13px; line-height: 1.6; color: #64748b; margin: 0 0 24px 0;">
                You received this email because notifications are enabled for this wallet.
            </p>
            
            <div style="font-size: 12px; color: #64748b; text-align: center; border-top: 1px solid #334155; padding-top: 16px;">
                Sent automatically by your Crypto Dashboard.
            </div>

        </div>
    </body>
    </html>
`))

func SubjectBalance(name string, balance float64) string {
	return fmt.Sprintf("Balance changed (%v) %s", balance, name)
}

func RenderBalanceEmail(data BalanceTemplateData) (string, error) {
	var buf bytes.Buffer
	if err := balanceEmailTmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
