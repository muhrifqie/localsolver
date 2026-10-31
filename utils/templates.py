"""Minimal HTML pages we serve to the browser to trigger each captcha and read its token.

We intercept the target URL with page.route() and fulfill it with one of these, so the
captcha widget runs under the real site origin without ever hitting the real server.
Keep the `<!-- cf turnstile -->` marker in HTML_TEMPLATE; solve_turnstile injects the div there.
"""

HTML_TEMPLATE = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>localsolver</title>
    <script src="https://challenges.cloudflare.com/turnstile/v0/api.js?onload=onloadTurnstileCallback" async defer></script>
</head>
<body>
    <!-- cf turnstile -->
    <p id="ip-display"></p>
</body>
</html>
"""

RECAPTCHA_V3_TEMPLATE = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>rc-v3</title>
    <script src="https://www.google.com/recaptcha/api.js?render={sitekey}"></script>
</head>
<body>
    <input type="hidden" id="recaptcha-token" name="recaptcha-token">
    <script>
        grecaptcha.ready(function() {{
            grecaptcha.execute('{sitekey}', {{action: '{action}'}}).then(function(token) {{
                document.getElementById('recaptcha-token').value = token;
            }});
        }});
    </script>
</body>
</html>
"""

RECAPTCHA_V2_TEMPLATE = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>rc-v2</title>
    <script src="https://www.google.com/recaptcha/api.js" async defer></script>
</head>
<body>
    <div class="g-recaptcha" data-sitekey="{sitekey}" data-callback="onSuccess"></div>
    <input type="hidden" id="recaptcha-response">
    <script>
        function onSuccess(token) {{ document.getElementById('recaptcha-response').value = token; }}
    </script>
</body>
</html>
"""

RECAPTCHA_V2_INVISIBLE_TEMPLATE = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>rc-v2-inv</title>
    <script src="https://www.google.com/recaptcha/api.js" async defer></script>
</head>
<body>
    <div class="g-recaptcha" data-sitekey="{sitekey}" data-size="invisible" data-callback="onSuccess"></div>
    <input type="hidden" id="recaptcha-response">
    <script>
        function onSuccess(token) {{ document.getElementById('recaptcha-response').value = token; }}
        setTimeout(function() {{ grecaptcha.execute(); }}, 1000);
    </script>
</body>
</html>
"""
