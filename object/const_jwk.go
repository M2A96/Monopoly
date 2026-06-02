package object

import (
	"reflect"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// JWKPubPrivKeys is a const.
var JWKPubPrivKeys = []map[string]string{
	// RSA JWK.
	{
		"alg": "RS512",
		"d":   `ds6MGqwaJgePFop_6EtTNxLz1mCeDULPfVFfYqSe2r97oi_n5vO3t_OSF331hcf0pv3y1fi0928UOc41ERn2a_h1X7SyvcaTZjNhhI9V5LphMnZBBpmVnh2FLjDfTt48W_qJSgT8IGnhkFStTEBsGtB2rMtXLDNdr4p7Z4RdBlRm8RTnWcKP-8IDxWsMdMrCA0TB9fw-mSzYXnSCQy3SJk9ywccoX-N2D_MlpTS53xkfAmdghlJJ-ceclSp-JtSNocO6IXzNcT9jr4Ptb-XWbkwldGScrVzNo5lMY1xi-piMnq5gwYB9LJa_Ji0AbP1SEkxM6aV8U3G83BrssaqgyD1BqBChGZ5TbtQ9wNZx4N0hCuGZHt9HpFoNbi-vPe_nOGTApG_Hj5t8Z6tqwfD8A1zEeePAfDfRMwbZ9nppn1A1FvnbLTIDrUW89X7HdrDbNEWYakCYrORb352xChp6GSiVkofYoYRS_aSh1a9QrsIvjQcCZapolBUfsHmaYd8hlSIwzCzggsEHgVF_PUXxW9nNGM8CTC7sO46KTHeEQTKl8tZUtwX3WdOqiZ45oATfTG3mgTU9xcfFVtq4rRP9R_Zm8_vA2fZvVKeb3TteccodiPQHWPU85Qo9AuIrrfOyEt9Kag6NOOmi4DmOxl4JkIvBK0rphZC3DPy57hqMZCE`,
		"dp":  `VN7X-15x3TGiTBGdNeSEPPYUfbyban2RbkqYIk1W-7pj7vJaDj5-_bBBHrIPuNewtYZiXlN34No-0VV9i9vDpJ6Vw7LBF6hzSzZUxTSUt9Re5KVxv5RlMbOca-7tN8nUPKHUz2JiJ7dSbNthXN1k7CU4ecDJ6P-6I8xUtFY1C83H5fCQlcFOOdS1l7e9SvX5X9dyEFckOcXV_ft3AQq6UGjmfXiNE1wfe3mjtxqHAYtUiZizCfCkdwSnG1ZHm8jVCMT0aT5jaNHA5CXqDYDYT4UJMbvwa4H9eRj8riRumd4iZ0Q-75sHaRNIOenqFzDld1UW3PTM0DlRLYITmGEAtQ`,
		"dq":  `EZAl9mwdPRvwuzFSN5thP_5sFcraw3W0kXJHTZ6MpQISwS8fRZt5QsTxPLcYYvz1fLAyh7BVEzUyi9JLizY4HIyuhQwKruNnAfJOqaGeWQlRnlCU96dHAVQL6MYLNjtE5eW0ouNhv8UjfVwcwJr2CxlBoDKDUbhJ3xIzym1DKorVpIwC0a4wSO8Gfz5zrZa4Qz2VGGBaooyRED119_QWX1ONI3DNL9fJYQUXCtfoL1M3XHceQS14C38ru_vhXzK2wWNSO8AvcPm4RsKNlAqkmQOZ6NgHU-YmePkVamIo4cC1loLijUMl4jnKQA0SdnVm0ic6FdSYf-jj-bRGqI6GkQ`,
		"e":   "AQAB",
		"kid": "16becd00-08a0-41b9-af1c-abca1329c8c5",
		"kty": "RSA",
		"n":   `mmpYq4VIwNn4HGX6U_gEPHegrpagA72y6YtBC8IcBYPpQ4VkZVqwFlCNEs4mR2Ou2EAkKvLa__rcQ0WRmnTEcnLy7DCrfdVw2rVtqWlzVAj3K55L72bHYV8ZCbzj7u3SEzSfuT0oW9JbrjteMTbqALaySdlvcmLolF6bz79c8S68E_Y2vjZkowqEXQDxlKYm-x6w-duht9f1_BXGCjxM8mbieowDSPcN3gLKTtjNIMq-TUiDgGZvGrFHWF_zaneYTMU1Ws-fFNHYWoe4x5oBe7ABoJrAM_RwlqNGSeI4aO2Oeb14BLe8-C5cgbTlCpbXbFBAm-UPw3kTmztNhJDY_TK4wMFkZYnlVgbE1Mey1oHYRAHb4xggIt1wvGDSNvbiXRD3Q5ltxC-2M467GMt-KsR0SFsFaHVf2TvkCJesvze4bP2XKT_QdUA8N7BfzMBi89NUJI2XC1FSo7ogidzyAZwtzjnrlh762N5tqqDRaNn5AMbXDUczp-oXqJz5sZkRwHnSYtWc2_scfYaUtZmk03jH63JFzp929PzxvhICAjWD5lhVExeKQAOdFNZrxL-18dfApjD_ZfyD30vcWzhvY6gK9C3RQDsuYCJ2eA7r1eVXmZmuENdSIibPvNE-lHSkgeRgvnNBSfp4qlimOaNqFU6GviIT4PT8xbUFvCLk_30`,
		"p":   `zK_Msk0vwoi_QfahA-D4HouN2zGoh3E3038b6ar2HkW_iW7RCJ2j1DSE3FNDwz_Utq3Uvxuy8O2NEAl8T3LCVdNK6qhM0iKol0Zyxz5AcSKGRj6Tk5shMx_rMdRpWYt78amH6lzAp9q2w74wRjrDxzGw5L-KODNt5fJTZiE2Od7Cvldy0sd-eaGF6S012xBNFDmFt2VkQ7K_WlvhcZypTUyn5oRnlwa3LHUvztg-0YCTT9ffng8FQOttHHuNicrWPk8R-gjmPHRhp582gMpWaSYClqrpvjC7xCUPKMPFRhQgFN0bfQg5d7nFYFthqoZ1nG19cojYC_cyA2JrdSS7fw`,
		"q":   `wSBHwYyzFKsINbGK3TDU2SgJHYGR98XtwGFHIDVb5oRI_PztTcbo5taX8_zIbCCqKiqTaO2mH14OednUmKUVFHrOP640s0oh56i-4W31djFAleomJwMI9dKF4HhwjHoe-e01oYEO6UXlTfxPlG7uvEalWH068QkWptUrftexTTrEhsWOgT73tx8-zxMGW9nD8zF_DaKp4AI3h6sMQmhjzDIEbH8q-w8sBL62ziEeNLh4QTkoj3OCnmCGfhfwq0ziR-gaE1jKycHbmLDPJR70IVcP_2VgYisODQnlWKMEoppCSTZWmYBOk76PXs-oLxIYYNFy_NnJNrIqCJn1jiOzAw`,
		"qi":  `fvinngALNDG-4dm8UiX5np-rbrSJwo_earGU9oWC_-YDEazsLaVZwSoTtDQ6N9ccoDyf0vnX0Y4iycsHf3sJxADiHWkHtybqxCF03BLezeSXP9Ss3GHyKm5FDLsE7mtj6q93njRKucEkeDXpGQjoV_wm398R_A2RHqdTGTFQQm4o-OCLNC561In88_djtJaY8P6vL_8SyGPG1T_-JOwbJnZc8Ns8fYAdIv4z3spjGobyi57-EZR6mefJxkYNx9ygq86TwWg1ZY6snaOp-ESFIzDK4H6PvedcUfd-kYyN45PfmGB6aXnVEcNNRbWmQY_tI-WrvX2dCIYWGlQjnHOYhA`,
		"use": "sig",
	},
	// HMAC HS256 JWK.
	{
		"alg": "HS256",
		"k":   "G37cfUp9nhwlxZDL2x0ecfKpzbhMT7zHYS786T-n0II",
		"kid": "27a7cb2b-6f0f-4722-a735-a45eb95b28a7",
		"kty": "oct",
	},
	// HMAC HS384 JWK.
	{
		"alg": "HS384",
		"k":   "AVQ-4XgHTI_KVV2h27nCBkTGb7N-K3QEghlB1sYYoNlXsEzKTv8YAXWdBp6cH4yc",
		"kid": "32502afd-077e-4c38-bb1c-9f7ee2069b0d",
		"kty": "oct",
	},
	// HMAC HS512 JWK.
	{
		"alg": "HS512",
		"k":   "_A3GhQMmfixjef5G9bFNKu7XhY7i1Tf5gyuWHrFIVTBk4t9APCX8Foq1SJWgCspLy3MuLgrI7js-0JS65M78dg",
		"kid": "1a35af02-71fe-4240-b9ed-f90482e405bc",
		"kty": "oct",
	},
	// RSA 2048-bit JWK.
	{
		"alg": "RS512",
		"d":   `ksDmucdMJXkFGZxiomNHnroOZxe8AmDLDGO1vhs-POa5PZM7mtUPonxwjVmthmpbZzla-kg55OFfO7YcXhg-Hm2OWTKwm73_rLh3JavaHjvBqsVKuorX3V3RYkSro6HyYIzFJ1Ek7sLxbjDRcDOj4ievSX0oN9l-JZhaDYlPlci5uJsoqro_YrE0PRRWVhtGynd-_aWgQv1YzkfZuMD-hJtDi1Im2humOWxA4eZrFs9eG-whXcOvaSwO4sSGbS99ecQZHM2TcdXeAs1PvjVgQ_dKnZlGN3lTWoWfQP55Z7Tgt8Nf1q4ZAKd-NlMe-7iqCFfsnFwXjSiaOa2CRGZn-Q`,
		"dp":  `lmmU_AG5SGxBhJqb8wxfNXDPJjf__i92BgJT2Vp4pskBbr5PGoyV0HbfUQVMnw977RONEurkR6O6gxZUeCclGt4kQlGZ-m0_XSWx13v9t9DIbheAtgVJ2mQyVDvK4m7aRYlEceFh0PsX8vYDS5o1txgPwb3oXkPTtrmbAGMUBpE`,
		"dq":  `mxRTU3QDyR2EnCv0Nl0TCF90oliJGAHR9HJmBe__EjuCBbwHfcT8OG3hWOv8vpzokQPRl5cQt3NckzX3fs6xlJN4Ai2Hh2zduKFVQ2p-AF2p6Yfahscjtq-GY9cB85NxLy2IXCC0PF--Sq9LOrTE9QV988SJy_yUrAjcZ5MmECk`,
		"e":   "AQAB",
		"kid": "cc34c0a0-bd5a-4a3c-a50d-a2a7db7643df",
		"kty": "RSA",
		"n":   `pjdss8ZaDfEH6K6U7GeW2nxDqR4IP049fk1fK0lndimbMMVBdPv_hSpm8T8EtBDxrUdi1OHZfMhUixGaut-3nQ4GG9nM249oxhCtxqqNvEXrmQRGqczyLxuh-fKn9Fg--hS9UpazHpfVAFnB5aCfXoNhPuI8oByyFKMKaOVgHNqP5NBEqabiLftZD3W_lsFCPGuzr4Vp0YS7zS2hDYScC2oOMu4rGU1LcMZf39p3153Cq7bS2Xh6Y-vw5pwzFYZdjQxDn8x8BG3fJ6j8TGLXQsbKH1218_HcUJRvMwdpbUQG5nvA2GXVqLqdwp054Lzk9_B_f1lVrmOKuHjTNHq48w`,
		"p":   `4A5nU4ahEww7B65yuzmGeCUUi8ikWzv1C81pSyUKvKzu8CX41hp9J6oRaLGesKImYiuVQK47FhZ--wwfpRwHvSxtNU9qXb8ewo-BvadyO1eVrIk4tNV543QlSe7pQAoJGkxCia5rfznAE3InKF4JvIlchyqs0RQ8wx7lULqwnn0`,
		"q":   `ven83GM6SfrmO-TBHbjTk6JhP_3CMsIvmSdo4KrbQNvp4vHO3w1_0zJ3URkmkYGhz2tgPlfd7v1l2I6QkIh4Bumdj6FyFZEBpxjE4MpfdNVcNINvVj87cLyTRmIcaGxmfylY7QErP8GFA-k4UoH_eQmGKGK44TRzYj5hZYGWIC8`,
		"qi":  `ldHXIrEmMZVaNwGzDF9WG8sHj2mOZmQpw9yrjLK9hAsmsNr5LTyqWAqJIYZSwPTYWhY4nu2O0EY9G9uYiqewXfCKw_UngrJt8Xwfq1Zruz0YY869zPN4GiE9-9rzdZB33RBw8kIOquY3MK74FMwCihYx_LiU2YTHkaoJ3ncvtvg`,
		"use": "sig",
	},
	// EC P-256 JWK.
	{
		"alg": "ES256",
		"crv": "P-256",
		"d":   "0g5vAEKzugrXaRbgKG0Tj2qJ5lMP4Bezds1_sTybkfk",
		"kid": "541934a7-8382-48b9-a43a-57d4cb4445d2",
		"kty": "EC",
		"x":   "SVqB4JcUD6lsfvqMr-OKUNUphdNn64Eay60978ZlL74",
		"y":   "lf0u0pMj4lGAzZix5u4Cm5CMQIgMNpkwy163wtKYVKI",
	},
	// Ed25519 JWK.
	{
		"alg": "EdDSA",
		"crv": "Ed25519",
		"d":   "nWGxne_9WmC6hEr0kuwsxERJxWl7MmkZcDusAxyuf2A",
		"kid": "8f760a1c-faf0-4e2d-9799-8a0ba78a7fb8",
		"kty": "OKP",
		"use": "sig",
		"x":   "11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo",
	},
	// HMAC HS256 JWK.
	{
		"alg": "HS256",
		"k":   "FdFYFzERwC2uCBB46pZQi4GG85LujR8obt-KWRBICVQ",
		"kid": "0afee142-a0af-4410-abcc-9f2d44ff45b5",
		"kty": "oct",
	},
	// AES 128-bit JWK.
	{
		"alg": "A128GCM",
		"k":   "c7WsUB6msAgIdDxTnT13Yw",
		"kid": "283e6c58-7962-411a-8468-8b26d6f8e89a",
		"kty": "oct",
	},
}

// KeyComparer is a function.
func KeyComparer(
	first jwk.Key,
	second jwk.Key,
) bool {
	return reflect.DeepEqual(first, second)
}

// SetComparer is a function.
func SetComparer(
	first jwk.Set,
	second jwk.Set,
) bool {
	return reflect.DeepEqual(first, second)
}
