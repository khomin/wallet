package notification

import (
	"html/template"
)

// Pre-parse template at package init level to catch syntax errors early and keep execution fast
var alertEmailTmpl = template.Must(template.New("alertEmail").Parse(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
	</head>
	<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #0f172a; color: #f8fafc; margin: 0; padding: 24px;">
		<div style="max-width: 560px; margin: 0 auto; background-color: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 32px;">
			
			<div style="display: inline-block; background-color: rgba(34, 197, 94, 0.1); color: #4ade80; border: 1px solid rgba(74, 222, 128, 0.2); font-size: 12px; font-weight: 600; padding: 4px 12px; border-radius: 9999px; text-transform: uppercase; margin-bottom: 16px;">
				Triggered
			</div>
			
			<h1 style="font-size: 20px; font-weight: 600; margin: 0 0 12px 0; color: #ffffff;">
				Price Alert for {{.CoinSymbol}}
			</h1>
			
			<p style="font-size: 14px; line-height: 1.6; color: #94a3b8; margin: 0 0 24px 0;">
				Hi {{.UserName}}, your target price alert was reached.
			</p>
			
			<!-- Bulletproof Table with inline styles and explicit cell widths -->
			<table width="100%" cellpadding="0" cellspacing="0" border="0" style="width: 100%; min-width: 100%; background-color: #0f172a; border: 1px solid #334155; border-radius: 8px; margin-bottom: 24px; border-collapse: collapse;">
				<tr>
					<td align="left" valign="middle" style="padding: 16px 20px; font-size: 15px; font-weight: 600; color: #94a3b8; text-align: left; width: 50%;">
						{{.CoinSymbol}} Current Price
					</td>
					<td align="right" valign="middle" style="padding: 16px 20px; font-size: 20px; font-weight: 700; color: #38bdf8; text-align: right; width: 50%;">
						${{.FormattedPrice}}
					</td>
				</tr>
			</table>
			
			<p style="font-size: 13px; line-height: 1.6; color: #64748b; margin: 0 0 24px 0;">
				This single-shot alert has been automatically completed and disabled.
			</p>
			
			<div style="font-size: 12px; color: #64748b; text-align: center; border-top: 1px solid #334155; padding-top: 16px;">
				Sent automatically by your Crypto Dashboard.
			</div>

		</div>
	</body>
	</html>
`))
