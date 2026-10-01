export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    
    // Origin target is VPS Caddy endpoint
    const origin = env.ORIGIN_URL || "https://8-231-102-98.sslip.io";
    const targetUrl = new URL(url.pathname + url.search, origin);
    
    const headers = new Headers(request.headers);
    headers.set("Host", new URL(origin).host);
    
    const forwardReq = new Request(targetUrl.toString(), {
      method: request.method,
      headers: headers,
      body: request.method !== "GET" && request.method !== "HEAD" ? request.body : undefined,
      redirect: "follow",
    });

    try {
      const response = await fetch(forwardReq);
      const responseHeaders = new Headers(response.headers);
      responseHeaders.set("Access-Control-Allow-Origin", "*");
      responseHeaders.set("Access-Control-Allow-Methods", "GET, HEAD, POST, OPTIONS");
      responseHeaders.set("Access-Control-Allow-Headers", "*");

      return new Response(response.body, {
        status: response.status,
        statusText: response.statusText,
        headers: responseHeaders,
      });
    } catch (err) {
      return new Response("Streaming Gateway Error: " + err.message, { status: 502 });
    }
  },
};
