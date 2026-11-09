import { urlBase64ToUint8Array } from "./web-push";

describe("urlBase64ToUint8Array", () => {
	it("decodes url-safe base64 without padding", () => {
		// "hi??" => aGk_Pw in url-safe base64
		expect(Array.from(urlBase64ToUint8Array("aGk_Pw"))).toEqual([
			104, 105, 63, 63,
		]);
	});
});
