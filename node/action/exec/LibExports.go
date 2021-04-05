package exec

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/futurehomeno/fimpgo"
	"github.com/thingsplex/tpflow/model"
	"go/constant"
	"go/token"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var Symbols = map[string]map[string]reflect.Value{}

func initScriptExports() {
	Symbols["github.com/thingsplex/tpflow/model"] = map[string]reflect.Value{
		"Message": reflect.ValueOf((*model.Message)(nil)),
		"Context": reflect.ValueOf((*model.Context)(nil)),
	}
	Symbols["github.com/thingsplex/tpflow/node/action/exec"] = map[string]reflect.Value{
		"ScriptParams": reflect.ValueOf((*ScriptParams)(nil)),
	}

	Symbols["github.com/futurehomeno/fimpgo"] = map[string]reflect.Value{
		"FimpMessage": reflect.ValueOf((*fimpgo.FimpMessage)(nil)),
		"NewMessage":   reflect.ValueOf(fimpgo.NewMessage),
		"Props": reflect.ValueOf((*fimpgo.Props)(nil)),
		"Tags": reflect.ValueOf((*fimpgo.Tags)(nil)),
		"Address": reflect.ValueOf((*fimpgo.Address)(nil)),
		"MqttTransport": reflect.ValueOf((*fimpgo.MqttTransport)(nil)),
	}



	Symbols["fmt"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Errorf":   reflect.ValueOf(fmt.Errorf),
		"Fprint":   reflect.ValueOf(fmt.Fprint),
		"Fprintf":  reflect.ValueOf(fmt.Fprintf),
		"Fprintln": reflect.ValueOf(fmt.Fprintln),
		"Fscan":    reflect.ValueOf(fmt.Fscan),
		"Fscanf":   reflect.ValueOf(fmt.Fscanf),
		"Fscanln":  reflect.ValueOf(fmt.Fscanln),
		"Print":    reflect.ValueOf(fmt.Print),
		"Printf":   reflect.ValueOf(fmt.Printf),
		"Println":  reflect.ValueOf(fmt.Println),
		"Sprint":   reflect.ValueOf(fmt.Sprint),
		"Sprintf":  reflect.ValueOf(fmt.Sprintf),
		"Sprintln": reflect.ValueOf(fmt.Sprintln),

		// type definitions
		"Formatter":  reflect.ValueOf((*fmt.Formatter)(nil)),
		"GoStringer": reflect.ValueOf((*fmt.GoStringer)(nil)),
		"ScanState":  reflect.ValueOf((*fmt.ScanState)(nil)),
		"Scanner":    reflect.ValueOf((*fmt.Scanner)(nil)),
		"State":      reflect.ValueOf((*fmt.State)(nil)),
		"Stringer":   reflect.ValueOf((*fmt.Stringer)(nil)),

	}

	Symbols["time"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"ANSIC":                  reflect.ValueOf(constant.MakeFromLiteral("\"Mon Jan _2 15:04:05 2006\"", token.STRING, 0)),
		"After":                  reflect.ValueOf(time.After),
		"AfterFunc":              reflect.ValueOf(time.AfterFunc),
		"April":                  reflect.ValueOf(time.April),
		"August":                 reflect.ValueOf(time.August),
		"Date":                   reflect.ValueOf(time.Date),
		"December":               reflect.ValueOf(time.December),
		"February":               reflect.ValueOf(time.February),
		"FixedZone":              reflect.ValueOf(time.FixedZone),
		"Friday":                 reflect.ValueOf(time.Friday),
		"Hour":                   reflect.ValueOf(time.Hour),
		"January":                reflect.ValueOf(time.January),
		"July":                   reflect.ValueOf(time.July),
		"June":                   reflect.ValueOf(time.June),
		"Kitchen":                reflect.ValueOf(constant.MakeFromLiteral("\"3:04PM\"", token.STRING, 0)),
		"LoadLocation":           reflect.ValueOf(time.LoadLocation),
		"LoadLocationFromTZData": reflect.ValueOf(time.LoadLocationFromTZData),
		"Local":                  reflect.ValueOf(&time.Local).Elem(),
		"March":                  reflect.ValueOf(time.March),
		"May":                    reflect.ValueOf(time.May),
		"Microsecond":            reflect.ValueOf(time.Microsecond),
		"Millisecond":            reflect.ValueOf(time.Millisecond),
		"Minute":                 reflect.ValueOf(time.Minute),
		"Monday":                 reflect.ValueOf(time.Monday),
		"Nanosecond":             reflect.ValueOf(time.Nanosecond),
		"NewTicker":              reflect.ValueOf(time.NewTicker),
		"NewTimer":               reflect.ValueOf(time.NewTimer),
		"November":               reflect.ValueOf(time.November),
		"Now":                    reflect.ValueOf(time.Now),
		"October":                reflect.ValueOf(time.October),
		"Parse":                  reflect.ValueOf(time.Parse),
		"ParseDuration":          reflect.ValueOf(time.ParseDuration),
		"ParseInLocation":        reflect.ValueOf(time.ParseInLocation),
		"RFC1123":                reflect.ValueOf(constant.MakeFromLiteral("\"Mon, 02 Jan 2006 15:04:05 MST\"", token.STRING, 0)),
		"RFC1123Z":               reflect.ValueOf(constant.MakeFromLiteral("\"Mon, 02 Jan 2006 15:04:05 -0700\"", token.STRING, 0)),
		"RFC3339":                reflect.ValueOf(constant.MakeFromLiteral("\"2006-01-02T15:04:05Z07:00\"", token.STRING, 0)),
		"RFC3339Nano":            reflect.ValueOf(constant.MakeFromLiteral("\"2006-01-02T15:04:05.999999999Z07:00\"", token.STRING, 0)),
		"RFC822":                 reflect.ValueOf(constant.MakeFromLiteral("\"02 Jan 06 15:04 MST\"", token.STRING, 0)),
		"RFC822Z":                reflect.ValueOf(constant.MakeFromLiteral("\"02 Jan 06 15:04 -0700\"", token.STRING, 0)),
		"RFC850":                 reflect.ValueOf(constant.MakeFromLiteral("\"Monday, 02-Jan-06 15:04:05 MST\"", token.STRING, 0)),
		"RubyDate":               reflect.ValueOf(constant.MakeFromLiteral("\"Mon Jan 02 15:04:05 -0700 2006\"", token.STRING, 0)),
		"Saturday":               reflect.ValueOf(time.Saturday),
		"Second":                 reflect.ValueOf(time.Second),
		"September":              reflect.ValueOf(time.September),
		"Since":                  reflect.ValueOf(time.Since),
		"Sleep":                  reflect.ValueOf(time.Sleep),
		"Stamp":                  reflect.ValueOf(constant.MakeFromLiteral("\"Jan _2 15:04:05\"", token.STRING, 0)),
		"StampMicro":             reflect.ValueOf(constant.MakeFromLiteral("\"Jan _2 15:04:05.000000\"", token.STRING, 0)),
		"StampMilli":             reflect.ValueOf(constant.MakeFromLiteral("\"Jan _2 15:04:05.000\"", token.STRING, 0)),
		"StampNano":              reflect.ValueOf(constant.MakeFromLiteral("\"Jan _2 15:04:05.000000000\"", token.STRING, 0)),
		"Sunday":                 reflect.ValueOf(time.Sunday),
		"Thursday":               reflect.ValueOf(time.Thursday),
		"Tick":                   reflect.ValueOf(time.Tick),
		"Tuesday":                reflect.ValueOf(time.Tuesday),
		"UTC":                    reflect.ValueOf(&time.UTC).Elem(),
		"Unix":                   reflect.ValueOf(time.Unix),
		"UnixDate":               reflect.ValueOf(constant.MakeFromLiteral("\"Mon Jan _2 15:04:05 MST 2006\"", token.STRING, 0)),
		"Until":                  reflect.ValueOf(time.Until),
		"Wednesday":              reflect.ValueOf(time.Wednesday),

		// type definitions
		"Duration":   reflect.ValueOf((*time.Duration)(nil)),
		"Location":   reflect.ValueOf((*time.Location)(nil)),
		"Month":      reflect.ValueOf((*time.Month)(nil)),
		"ParseError": reflect.ValueOf((*time.ParseError)(nil)),
		"Ticker":     reflect.ValueOf((*time.Ticker)(nil)),
		"Time":       reflect.ValueOf((*time.Time)(nil)),
		"Timer":      reflect.ValueOf((*time.Timer)(nil)),
		"Weekday":    reflect.ValueOf((*time.Weekday)(nil)),
	}

	Symbols["strconv"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"AppendBool":               reflect.ValueOf(strconv.AppendBool),
		"AppendFloat":              reflect.ValueOf(strconv.AppendFloat),
		"AppendInt":                reflect.ValueOf(strconv.AppendInt),
		"AppendQuote":              reflect.ValueOf(strconv.AppendQuote),
		"AppendQuoteRune":          reflect.ValueOf(strconv.AppendQuoteRune),
		"AppendQuoteRuneToASCII":   reflect.ValueOf(strconv.AppendQuoteRuneToASCII),
		"AppendQuoteRuneToGraphic": reflect.ValueOf(strconv.AppendQuoteRuneToGraphic),
		"AppendQuoteToASCII":       reflect.ValueOf(strconv.AppendQuoteToASCII),
		"AppendQuoteToGraphic":     reflect.ValueOf(strconv.AppendQuoteToGraphic),
		"AppendUint":               reflect.ValueOf(strconv.AppendUint),
		"Atoi":                     reflect.ValueOf(strconv.Atoi),
		"CanBackquote":             reflect.ValueOf(strconv.CanBackquote),
		"ErrRange":                 reflect.ValueOf(&strconv.ErrRange).Elem(),
		"ErrSyntax":                reflect.ValueOf(&strconv.ErrSyntax).Elem(),
		"FormatBool":               reflect.ValueOf(strconv.FormatBool),
		"FormatComplex":            reflect.ValueOf(strconv.FormatComplex),
		"FormatFloat":              reflect.ValueOf(strconv.FormatFloat),
		"FormatInt":                reflect.ValueOf(strconv.FormatInt),
		"FormatUint":               reflect.ValueOf(strconv.FormatUint),
		"IntSize":                  reflect.ValueOf(constant.MakeFromLiteral("64", token.INT, 0)),
		"IsGraphic":                reflect.ValueOf(strconv.IsGraphic),
		"IsPrint":                  reflect.ValueOf(strconv.IsPrint),
		"Itoa":                     reflect.ValueOf(strconv.Itoa),
		"ParseBool":                reflect.ValueOf(strconv.ParseBool),
		"ParseComplex":             reflect.ValueOf(strconv.ParseComplex),
		"ParseFloat":               reflect.ValueOf(strconv.ParseFloat),
		"ParseInt":                 reflect.ValueOf(strconv.ParseInt),
		"ParseUint":                reflect.ValueOf(strconv.ParseUint),
		"Quote":                    reflect.ValueOf(strconv.Quote),
		"QuoteRune":                reflect.ValueOf(strconv.QuoteRune),
		"QuoteRuneToASCII":         reflect.ValueOf(strconv.QuoteRuneToASCII),
		"QuoteRuneToGraphic":       reflect.ValueOf(strconv.QuoteRuneToGraphic),
		"QuoteToASCII":             reflect.ValueOf(strconv.QuoteToASCII),
		"QuoteToGraphic":           reflect.ValueOf(strconv.QuoteToGraphic),
		"Unquote":                  reflect.ValueOf(strconv.Unquote),
		"UnquoteChar":              reflect.ValueOf(strconv.UnquoteChar),

		// type definitions
		"NumError": reflect.ValueOf((*strconv.NumError)(nil)),
	}

	Symbols["strings"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Compare":        reflect.ValueOf(strings.Compare),
		"Contains":       reflect.ValueOf(strings.Contains),
		"ContainsAny":    reflect.ValueOf(strings.ContainsAny),
		"ContainsRune":   reflect.ValueOf(strings.ContainsRune),
		"Count":          reflect.ValueOf(strings.Count),
		"EqualFold":      reflect.ValueOf(strings.EqualFold),
		"Fields":         reflect.ValueOf(strings.Fields),
		"FieldsFunc":     reflect.ValueOf(strings.FieldsFunc),
		"HasPrefix":      reflect.ValueOf(strings.HasPrefix),
		"HasSuffix":      reflect.ValueOf(strings.HasSuffix),
		"Index":          reflect.ValueOf(strings.Index),
		"IndexAny":       reflect.ValueOf(strings.IndexAny),
		"IndexByte":      reflect.ValueOf(strings.IndexByte),
		"IndexFunc":      reflect.ValueOf(strings.IndexFunc),
		"IndexRune":      reflect.ValueOf(strings.IndexRune),
		"Join":           reflect.ValueOf(strings.Join),
		"LastIndex":      reflect.ValueOf(strings.LastIndex),
		"LastIndexAny":   reflect.ValueOf(strings.LastIndexAny),
		"LastIndexByte":  reflect.ValueOf(strings.LastIndexByte),
		"LastIndexFunc":  reflect.ValueOf(strings.LastIndexFunc),
		"Map":            reflect.ValueOf(strings.Map),
		"NewReader":      reflect.ValueOf(strings.NewReader),
		"NewReplacer":    reflect.ValueOf(strings.NewReplacer),
		"Repeat":         reflect.ValueOf(strings.Repeat),
		"Replace":        reflect.ValueOf(strings.Replace),
		"ReplaceAll":     reflect.ValueOf(strings.ReplaceAll),
		"Split":          reflect.ValueOf(strings.Split),
		"SplitAfter":     reflect.ValueOf(strings.SplitAfter),
		"SplitAfterN":    reflect.ValueOf(strings.SplitAfterN),
		"SplitN":         reflect.ValueOf(strings.SplitN),
		"Title":          reflect.ValueOf(strings.Title),
		"ToLower":        reflect.ValueOf(strings.ToLower),
		"ToLowerSpecial": reflect.ValueOf(strings.ToLowerSpecial),
		"ToTitle":        reflect.ValueOf(strings.ToTitle),
		"ToTitleSpecial": reflect.ValueOf(strings.ToTitleSpecial),
		"ToUpper":        reflect.ValueOf(strings.ToUpper),
		"ToUpperSpecial": reflect.ValueOf(strings.ToUpperSpecial),
		"ToValidUTF8":    reflect.ValueOf(strings.ToValidUTF8),
		"Trim":           reflect.ValueOf(strings.Trim),
		"TrimFunc":       reflect.ValueOf(strings.TrimFunc),
		"TrimLeft":       reflect.ValueOf(strings.TrimLeft),
		"TrimLeftFunc":   reflect.ValueOf(strings.TrimLeftFunc),
		"TrimPrefix":     reflect.ValueOf(strings.TrimPrefix),
		"TrimRight":      reflect.ValueOf(strings.TrimRight),
		"TrimRightFunc":  reflect.ValueOf(strings.TrimRightFunc),
		"TrimSpace":      reflect.ValueOf(strings.TrimSpace),
		"TrimSuffix":     reflect.ValueOf(strings.TrimSuffix),

		// type definitions
		"Builder":  reflect.ValueOf((*strings.Builder)(nil)),
		"Reader":   reflect.ValueOf((*strings.Reader)(nil)),
		"Replacer": reflect.ValueOf((*strings.Replacer)(nil)),
	}

	Symbols["math"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Abs":                    reflect.ValueOf(math.Abs),
		"Acos":                   reflect.ValueOf(math.Acos),
		"Acosh":                  reflect.ValueOf(math.Acosh),
		"Asin":                   reflect.ValueOf(math.Asin),
		"Asinh":                  reflect.ValueOf(math.Asinh),
		"Atan":                   reflect.ValueOf(math.Atan),
		"Atan2":                  reflect.ValueOf(math.Atan2),
		"Atanh":                  reflect.ValueOf(math.Atanh),
		"Cbrt":                   reflect.ValueOf(math.Cbrt),
		"Ceil":                   reflect.ValueOf(math.Ceil),
		"Copysign":               reflect.ValueOf(math.Copysign),
		"Cos":                    reflect.ValueOf(math.Cos),
		"Cosh":                   reflect.ValueOf(math.Cosh),
		"Dim":                    reflect.ValueOf(math.Dim),
		"E":                      reflect.ValueOf(constant.MakeFromLiteral("2.71828182845904523536028747135266249775724709369995957496696762566337824315673231520670375558666729784504486779277967997696994772644702281675346915668215131895555530285035761295375777990557253360748291015625", token.FLOAT, 0)),
		"Erf":                    reflect.ValueOf(math.Erf),
		"Erfc":                   reflect.ValueOf(math.Erfc),
		"Erfcinv":                reflect.ValueOf(math.Erfcinv),
		"Erfinv":                 reflect.ValueOf(math.Erfinv),
		"Exp":                    reflect.ValueOf(math.Exp),
		"Exp2":                   reflect.ValueOf(math.Exp2),
		"Expm1":                  reflect.ValueOf(math.Expm1),
		"FMA":                    reflect.ValueOf(math.FMA),
		"Float32bits":            reflect.ValueOf(math.Float32bits),
		"Float32frombits":        reflect.ValueOf(math.Float32frombits),
		"Float64bits":            reflect.ValueOf(math.Float64bits),
		"Float64frombits":        reflect.ValueOf(math.Float64frombits),
		"Floor":                  reflect.ValueOf(math.Floor),
		"Frexp":                  reflect.ValueOf(math.Frexp),
		"Gamma":                  reflect.ValueOf(math.Gamma),
		"Hypot":                  reflect.ValueOf(math.Hypot),
		"Ilogb":                  reflect.ValueOf(math.Ilogb),
		"Inf":                    reflect.ValueOf(math.Inf),
		"IsInf":                  reflect.ValueOf(math.IsInf),
		"IsNaN":                  reflect.ValueOf(math.IsNaN),
		"J0":                     reflect.ValueOf(math.J0),
		"J1":                     reflect.ValueOf(math.J1),
		"Jn":                     reflect.ValueOf(math.Jn),
		"Ldexp":                  reflect.ValueOf(math.Ldexp),
		"Lgamma":                 reflect.ValueOf(math.Lgamma),
		"Ln10":                   reflect.ValueOf(constant.MakeFromLiteral("2.30258509299404568401799145468436420760110148862877297603332784146804725494827975466552490443295866962642372461496758838959542646932914211937012833592062802600362869664962772731087170541286468505859375", token.FLOAT, 0)),
		"Ln2":                    reflect.ValueOf(constant.MakeFromLiteral("0.6931471805599453094172321214581765680755001343602552541206800092715999496201383079363438206637927920954189307729314303884387720696314608777673678644642390655170150035209453154294578780536539852619171142578125", token.FLOAT, 0)),
		"Log":                    reflect.ValueOf(math.Log),
		"Log10":                  reflect.ValueOf(math.Log10),
		"Log10E":                 reflect.ValueOf(constant.MakeFromLiteral("0.43429448190325182765112891891660508229439700580366656611445378416636798190620320263064286300825210972160277489744884502676719847561509639618196799746596688688378591625127711495224502868950366973876953125", token.FLOAT, 0)),
		"Log1p":                  reflect.ValueOf(math.Log1p),
		"Log2":                   reflect.ValueOf(math.Log2),
		"Log2E":                  reflect.ValueOf(constant.MakeFromLiteral("1.44269504088896340735992468100189213742664595415298593413544940772066427768997545329060870636212628972710992130324953463427359402479619301286929040235571747101382214539290471666532766903401352465152740478515625", token.FLOAT, 0)),
		"Logb":                   reflect.ValueOf(math.Logb),
		"Max":                    reflect.ValueOf(math.Max),
		"MaxFloat32":             reflect.ValueOf(constant.MakeFromLiteral("340282346638528859811704183484516925440", token.FLOAT, 0)),
		"MaxFloat64":             reflect.ValueOf(constant.MakeFromLiteral("179769313486231570814527423731704356798100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", token.FLOAT, 0)),
		"MaxInt16":               reflect.ValueOf(constant.MakeFromLiteral("32767", token.INT, 0)),
		"MaxInt32":               reflect.ValueOf(constant.MakeFromLiteral("2147483647", token.INT, 0)),
		"MaxInt64":               reflect.ValueOf(constant.MakeFromLiteral("9223372036854775807", token.INT, 0)),
		"MaxInt8":                reflect.ValueOf(constant.MakeFromLiteral("127", token.INT, 0)),
		"MaxUint16":              reflect.ValueOf(constant.MakeFromLiteral("65535", token.INT, 0)),
		"MaxUint32":              reflect.ValueOf(constant.MakeFromLiteral("4294967295", token.INT, 0)),
		"MaxUint64":              reflect.ValueOf(constant.MakeFromLiteral("18446744073709551615", token.INT, 0)),
		"MaxUint8":               reflect.ValueOf(constant.MakeFromLiteral("255", token.INT, 0)),
		"Min":                    reflect.ValueOf(math.Min),
		"MinInt16":               reflect.ValueOf(constant.MakeFromLiteral("-32768", token.INT, 0)),
		"MinInt32":               reflect.ValueOf(constant.MakeFromLiteral("-2147483648", token.INT, 0)),
		"MinInt64":               reflect.ValueOf(constant.MakeFromLiteral("-9223372036854775808", token.INT, 0)),
		"MinInt8":                reflect.ValueOf(constant.MakeFromLiteral("-128", token.INT, 0)),
		"Mod":                    reflect.ValueOf(math.Mod),
		"Modf":                   reflect.ValueOf(math.Modf),
		"NaN":                    reflect.ValueOf(math.NaN),
		"Nextafter":              reflect.ValueOf(math.Nextafter),
		"Nextafter32":            reflect.ValueOf(math.Nextafter32),
		"Phi":                    reflect.ValueOf(constant.MakeFromLiteral("1.6180339887498948482045868343656381177203091798057628621354486119746080982153796619881086049305501566952211682590824739205931370737029882996587050475921915678674035433959321750307935872115194797515869140625", token.FLOAT, 0)),
		"Pi":                     reflect.ValueOf(constant.MakeFromLiteral("3.141592653589793238462643383279502884197169399375105820974944594789982923695635954704435713335896673485663389728754819466702315787113662862838515639906529162340867271374644786874341662041842937469482421875", token.FLOAT, 0)),
		"Pow":                    reflect.ValueOf(math.Pow),
		"Pow10":                  reflect.ValueOf(math.Pow10),
		"Remainder":              reflect.ValueOf(math.Remainder),
		"Round":                  reflect.ValueOf(math.Round),
		"RoundToEven":            reflect.ValueOf(math.RoundToEven),
		"Signbit":                reflect.ValueOf(math.Signbit),
		"Sin":                    reflect.ValueOf(math.Sin),
		"Sincos":                 reflect.ValueOf(math.Sincos),
		"Sinh":                   reflect.ValueOf(math.Sinh),
		"SmallestNonzeroFloat32": reflect.ValueOf(constant.MakeFromLiteral("1.40129846432481707092372958328991613128000000000000000000000000000000000000000000001246655487714533538006789189734126694785975183981128816138510360971472225738624150874949653910667523779981133927289771669016713539217953030564201688027906006008453304556102801950542906382507e-45", token.FLOAT, 0)),
		"SmallestNonzeroFloat64": reflect.ValueOf(constant.MakeFromLiteral("4.94065645841246544176568792868221372365099999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999916206614696136086629714037163874026187912451674985660337336755242863513549746484310667379088263176934591818322489862214324814281481943599945502119376688748731948897748561110123901991443297110206447991752071007926740839424145013355231935665542622515363894390826799291671723318261174778903704064716351336223785714389641180220184242018383103204287325861250404139399888498504162666394779407509786431980433771341978183418568838015304951087487907666317075235615216699116844779095660202193409146032665221882798856203896125090454090026556150624798681464913851491093798848436664885581161128190046248588053014958829424991704801027040654863867512297941601850496672190315253109308532379657238854928816482120688440415705411555019932096150435627305446214567713171657554140575630917301482608119551500514805985376055777894871863446222606532650275466165274006e-324", token.FLOAT, 0)),
		"Sqrt":                   reflect.ValueOf(math.Sqrt),
		"Sqrt2":                  reflect.ValueOf(constant.MakeFromLiteral("1.414213562373095048801688724209698078569671875376948073176679739576083351575381440094441524123797447886801949755143139115339040409162552642832693297721230919563348109313505318596071447245776653289794921875", token.FLOAT, 0)),
		"SqrtE":                  reflect.ValueOf(constant.MakeFromLiteral("1.64872127070012814684865078781416357165377610071014801157507931167328763229187870850146925823776361770041160388013884200789716007979526823569827080974091691342077871211546646890155898290686309337615966796875", token.FLOAT, 0)),
		"SqrtPhi":                reflect.ValueOf(constant.MakeFromLiteral("1.2720196495140689642524224617374914917156080418400962486166403754616080542166459302584536396369727769747312116100875915825863540562126478288118732191412003988041797518382391984914647764526307582855224609375", token.FLOAT, 0)),
		"SqrtPi":                 reflect.ValueOf(constant.MakeFromLiteral("1.772453850905516027298167483341145182797549456122387128213807789740599698370237052541269446184448945647349951047154197675245574635259260134350885938555625028620527962319730619356050738133490085601806640625", token.FLOAT, 0)),
		"Tan":                    reflect.ValueOf(math.Tan),
		"Tanh":                   reflect.ValueOf(math.Tanh),
		"Trunc":                  reflect.ValueOf(math.Trunc),
		"Y0":                     reflect.ValueOf(math.Y0),
		"Y1":                     reflect.ValueOf(math.Y1),
		"Yn":                     reflect.ValueOf(math.Yn),
	}

	Symbols["sort"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Float64s":          reflect.ValueOf(sort.Float64s),
		"Float64sAreSorted": reflect.ValueOf(sort.Float64sAreSorted),
		"Ints":              reflect.ValueOf(sort.Ints),
		"IntsAreSorted":     reflect.ValueOf(sort.IntsAreSorted),
		"IsSorted":          reflect.ValueOf(sort.IsSorted),
		"Reverse":           reflect.ValueOf(sort.Reverse),
		"Search":            reflect.ValueOf(sort.Search),
		"SearchFloat64s":    reflect.ValueOf(sort.SearchFloat64s),
		"SearchInts":        reflect.ValueOf(sort.SearchInts),
		"SearchStrings":     reflect.ValueOf(sort.SearchStrings),
		"Slice":             reflect.ValueOf(sort.Slice),
		"SliceIsSorted":     reflect.ValueOf(sort.SliceIsSorted),
		"SliceStable":       reflect.ValueOf(sort.SliceStable),
		"Sort":              reflect.ValueOf(sort.Sort),
		"Stable":            reflect.ValueOf(sort.Stable),
		"Strings":           reflect.ValueOf(sort.Strings),
		"StringsAreSorted":  reflect.ValueOf(sort.StringsAreSorted),

		// type definitions
		"Float64Slice": reflect.ValueOf((*sort.Float64Slice)(nil)),
		"IntSlice":     reflect.ValueOf((*sort.IntSlice)(nil)),
		"Interface":    reflect.ValueOf((*sort.Interface)(nil)),
		"StringSlice":  reflect.ValueOf((*sort.StringSlice)(nil)),

		// interface wrapper definitions
		"_Interface": reflect.ValueOf((*_sort_Interface)(nil)),
	}

	Symbols["encoding/json"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Compact":       reflect.ValueOf(json.Compact),
		"HTMLEscape":    reflect.ValueOf(json.HTMLEscape),
		"Indent":        reflect.ValueOf(json.Indent),
		"Marshal":       reflect.ValueOf(json.Marshal),
		"MarshalIndent": reflect.ValueOf(json.MarshalIndent),
		"NewDecoder":    reflect.ValueOf(json.NewDecoder),
		"NewEncoder":    reflect.ValueOf(json.NewEncoder),
		"Unmarshal":     reflect.ValueOf(json.Unmarshal),
		"Valid":         reflect.ValueOf(json.Valid),

		// type definitions
		"Decoder":               reflect.ValueOf((*json.Decoder)(nil)),
		"Delim":                 reflect.ValueOf((*json.Delim)(nil)),
		"Encoder":               reflect.ValueOf((*json.Encoder)(nil)),
		"InvalidUTF8Error":      reflect.ValueOf((*json.InvalidUTF8Error)(nil)),
		"InvalidUnmarshalError": reflect.ValueOf((*json.InvalidUnmarshalError)(nil)),
		"Marshaler":             reflect.ValueOf((*json.Marshaler)(nil)),
		"MarshalerError":        reflect.ValueOf((*json.MarshalerError)(nil)),
		"Number":                reflect.ValueOf((*json.Number)(nil)),
		"RawMessage":            reflect.ValueOf((*json.RawMessage)(nil)),
		"SyntaxError":           reflect.ValueOf((*json.SyntaxError)(nil)),
		"Token":                 reflect.ValueOf((*json.Token)(nil)),
		"UnmarshalFieldError":   reflect.ValueOf((*json.UnmarshalFieldError)(nil)),
		"UnmarshalTypeError":    reflect.ValueOf((*json.UnmarshalTypeError)(nil)),
		"Unmarshaler":           reflect.ValueOf((*json.Unmarshaler)(nil)),
		"UnsupportedTypeError":  reflect.ValueOf((*json.UnsupportedTypeError)(nil)),
		"UnsupportedValueError": reflect.ValueOf((*json.UnsupportedValueError)(nil)),

		// interface wrapper definitions
		"_Marshaler":   reflect.ValueOf((*_encoding_json_Marshaler)(nil)),
		"_Token":       reflect.ValueOf((*_encoding_json_Token)(nil)),
		"_Unmarshaler": reflect.ValueOf((*_encoding_json_Unmarshaler)(nil)),
	}

	Symbols["net"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"CIDRMask":                   reflect.ValueOf(net.CIDRMask),
		"DefaultResolver":            reflect.ValueOf(&net.DefaultResolver).Elem(),
		"Dial":                       reflect.ValueOf(net.Dial),
		"DialIP":                     reflect.ValueOf(net.DialIP),
		"DialTCP":                    reflect.ValueOf(net.DialTCP),
		"DialTimeout":                reflect.ValueOf(net.DialTimeout),
		"DialUDP":                    reflect.ValueOf(net.DialUDP),
		"DialUnix":                   reflect.ValueOf(net.DialUnix),
		"ErrWriteToConnected":        reflect.ValueOf(&net.ErrWriteToConnected).Elem(),
		"FileConn":                   reflect.ValueOf(net.FileConn),
		"FileListener":               reflect.ValueOf(net.FileListener),
		"FilePacketConn":             reflect.ValueOf(net.FilePacketConn),
		"FlagBroadcast":              reflect.ValueOf(net.FlagBroadcast),
		"FlagLoopback":               reflect.ValueOf(net.FlagLoopback),
		"FlagMulticast":              reflect.ValueOf(net.FlagMulticast),
		"FlagPointToPoint":           reflect.ValueOf(net.FlagPointToPoint),
		"FlagUp":                     reflect.ValueOf(net.FlagUp),
		"IPv4":                       reflect.ValueOf(net.IPv4),
		"IPv4Mask":                   reflect.ValueOf(net.IPv4Mask),
		"IPv4allrouter":              reflect.ValueOf(&net.IPv4allrouter).Elem(),
		"IPv4allsys":                 reflect.ValueOf(&net.IPv4allsys).Elem(),
		"IPv4bcast":                  reflect.ValueOf(&net.IPv4bcast).Elem(),
		"IPv4len":                    reflect.ValueOf(constant.MakeFromLiteral("4", token.INT, 0)),
		"IPv4zero":                   reflect.ValueOf(&net.IPv4zero).Elem(),
		"IPv6interfacelocalallnodes": reflect.ValueOf(&net.IPv6interfacelocalallnodes).Elem(),
		"IPv6len":                    reflect.ValueOf(constant.MakeFromLiteral("16", token.INT, 0)),
		"IPv6linklocalallnodes":      reflect.ValueOf(&net.IPv6linklocalallnodes).Elem(),
		"IPv6linklocalallrouters":    reflect.ValueOf(&net.IPv6linklocalallrouters).Elem(),
		"IPv6loopback":               reflect.ValueOf(&net.IPv6loopback).Elem(),
		"IPv6unspecified":            reflect.ValueOf(&net.IPv6unspecified).Elem(),
		"IPv6zero":                   reflect.ValueOf(&net.IPv6zero).Elem(),
		"InterfaceAddrs":             reflect.ValueOf(net.InterfaceAddrs),
		"InterfaceByIndex":           reflect.ValueOf(net.InterfaceByIndex),
		"InterfaceByName":            reflect.ValueOf(net.InterfaceByName),
		"Interfaces":                 reflect.ValueOf(net.Interfaces),
		"JoinHostPort":               reflect.ValueOf(net.JoinHostPort),
		"Listen":                     reflect.ValueOf(net.Listen),
		"ListenIP":                   reflect.ValueOf(net.ListenIP),
		"ListenMulticastUDP":         reflect.ValueOf(net.ListenMulticastUDP),
		"ListenPacket":               reflect.ValueOf(net.ListenPacket),
		"ListenTCP":                  reflect.ValueOf(net.ListenTCP),
		"ListenUDP":                  reflect.ValueOf(net.ListenUDP),
		"ListenUnix":                 reflect.ValueOf(net.ListenUnix),
		"ListenUnixgram":             reflect.ValueOf(net.ListenUnixgram),
		"LookupAddr":                 reflect.ValueOf(net.LookupAddr),
		"LookupCNAME":                reflect.ValueOf(net.LookupCNAME),
		"LookupHost":                 reflect.ValueOf(net.LookupHost),
		"LookupIP":                   reflect.ValueOf(net.LookupIP),
		"LookupMX":                   reflect.ValueOf(net.LookupMX),
		"LookupNS":                   reflect.ValueOf(net.LookupNS),
		"LookupPort":                 reflect.ValueOf(net.LookupPort),
		"LookupSRV":                  reflect.ValueOf(net.LookupSRV),
		"LookupTXT":                  reflect.ValueOf(net.LookupTXT),
		"ParseCIDR":                  reflect.ValueOf(net.ParseCIDR),
		"ParseIP":                    reflect.ValueOf(net.ParseIP),
		"ParseMAC":                   reflect.ValueOf(net.ParseMAC),
		"Pipe":                       reflect.ValueOf(net.Pipe),
		"ResolveIPAddr":              reflect.ValueOf(net.ResolveIPAddr),
		"ResolveTCPAddr":             reflect.ValueOf(net.ResolveTCPAddr),
		"ResolveUDPAddr":             reflect.ValueOf(net.ResolveUDPAddr),
		"ResolveUnixAddr":            reflect.ValueOf(net.ResolveUnixAddr),
		"SplitHostPort":              reflect.ValueOf(net.SplitHostPort),

		// type definitions
		"Addr":                reflect.ValueOf((*net.Addr)(nil)),
		"AddrError":           reflect.ValueOf((*net.AddrError)(nil)),
		"Buffers":             reflect.ValueOf((*net.Buffers)(nil)),
		"Conn":                reflect.ValueOf((*net.Conn)(nil)),
		"DNSConfigError":      reflect.ValueOf((*net.DNSConfigError)(nil)),
		"DNSError":            reflect.ValueOf((*net.DNSError)(nil)),
		"Dialer":              reflect.ValueOf((*net.Dialer)(nil)),
		"Error":               reflect.ValueOf((*net.Error)(nil)),
		"Flags":               reflect.ValueOf((*net.Flags)(nil)),
		"HardwareAddr":        reflect.ValueOf((*net.HardwareAddr)(nil)),
		"IP":                  reflect.ValueOf((*net.IP)(nil)),
		"IPAddr":              reflect.ValueOf((*net.IPAddr)(nil)),
		"IPConn":              reflect.ValueOf((*net.IPConn)(nil)),
		"IPMask":              reflect.ValueOf((*net.IPMask)(nil)),
		"IPNet":               reflect.ValueOf((*net.IPNet)(nil)),
		"Interface":           reflect.ValueOf((*net.Interface)(nil)),
		"InvalidAddrError":    reflect.ValueOf((*net.InvalidAddrError)(nil)),
		"ListenConfig":        reflect.ValueOf((*net.ListenConfig)(nil)),
		"Listener":            reflect.ValueOf((*net.Listener)(nil)),
		"MX":                  reflect.ValueOf((*net.MX)(nil)),
		"NS":                  reflect.ValueOf((*net.NS)(nil)),
		"OpError":             reflect.ValueOf((*net.OpError)(nil)),
		"PacketConn":          reflect.ValueOf((*net.PacketConn)(nil)),
		"ParseError":          reflect.ValueOf((*net.ParseError)(nil)),
		"Resolver":            reflect.ValueOf((*net.Resolver)(nil)),
		"SRV":                 reflect.ValueOf((*net.SRV)(nil)),
		"TCPAddr":             reflect.ValueOf((*net.TCPAddr)(nil)),
		"TCPConn":             reflect.ValueOf((*net.TCPConn)(nil)),
		"TCPListener":         reflect.ValueOf((*net.TCPListener)(nil)),
		"UDPAddr":             reflect.ValueOf((*net.UDPAddr)(nil)),
		"UDPConn":             reflect.ValueOf((*net.UDPConn)(nil)),
		"UnixAddr":            reflect.ValueOf((*net.UnixAddr)(nil)),
		"UnixConn":            reflect.ValueOf((*net.UnixConn)(nil)),
		"UnixListener":        reflect.ValueOf((*net.UnixListener)(nil)),
		"UnknownNetworkError": reflect.ValueOf((*net.UnknownNetworkError)(nil)),

		// interface wrapper definitions
		"_Addr":       reflect.ValueOf((*_net_Addr)(nil)),
		"_Conn":       reflect.ValueOf((*_net_Conn)(nil)),
		"_Error":      reflect.ValueOf((*_net_Error)(nil)),
		"_Listener":   reflect.ValueOf((*_net_Listener)(nil)),
		"_PacketConn": reflect.ValueOf((*_net_PacketConn)(nil)),
	}

	Symbols["net/http"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"CanonicalHeaderKey":                  reflect.ValueOf(http.CanonicalHeaderKey),
		"DefaultClient":                       reflect.ValueOf(&http.DefaultClient).Elem(),
		"DefaultMaxHeaderBytes":               reflect.ValueOf(constant.MakeFromLiteral("1048576", token.INT, 0)),
		"DefaultMaxIdleConnsPerHost":          reflect.ValueOf(constant.MakeFromLiteral("2", token.INT, 0)),
		"DefaultServeMux":                     reflect.ValueOf(&http.DefaultServeMux).Elem(),
		"DefaultTransport":                    reflect.ValueOf(&http.DefaultTransport).Elem(),
		"DetectContentType":                   reflect.ValueOf(http.DetectContentType),
		"ErrAbortHandler":                     reflect.ValueOf(&http.ErrAbortHandler).Elem(),
		"ErrBodyNotAllowed":                   reflect.ValueOf(&http.ErrBodyNotAllowed).Elem(),
		"ErrBodyReadAfterClose":               reflect.ValueOf(&http.ErrBodyReadAfterClose).Elem(),
		"ErrContentLength":                    reflect.ValueOf(&http.ErrContentLength).Elem(),
		"ErrHandlerTimeout":                   reflect.ValueOf(&http.ErrHandlerTimeout).Elem(),
		"ErrHeaderTooLong":                    reflect.ValueOf(&http.ErrHeaderTooLong).Elem(),
		"ErrHijacked":                         reflect.ValueOf(&http.ErrHijacked).Elem(),
		"ErrLineTooLong":                      reflect.ValueOf(&http.ErrLineTooLong).Elem(),
		"ErrMissingBoundary":                  reflect.ValueOf(&http.ErrMissingBoundary).Elem(),
		"ErrMissingContentLength":             reflect.ValueOf(&http.ErrMissingContentLength).Elem(),
		"ErrMissingFile":                      reflect.ValueOf(&http.ErrMissingFile).Elem(),
		"ErrNoCookie":                         reflect.ValueOf(&http.ErrNoCookie).Elem(),
		"ErrNoLocation":                       reflect.ValueOf(&http.ErrNoLocation).Elem(),
		"ErrNotMultipart":                     reflect.ValueOf(&http.ErrNotMultipart).Elem(),
		"ErrNotSupported":                     reflect.ValueOf(&http.ErrNotSupported).Elem(),
		"ErrServerClosed":                     reflect.ValueOf(&http.ErrServerClosed).Elem(),
		"ErrShortBody":                        reflect.ValueOf(&http.ErrShortBody).Elem(),
		"ErrSkipAltProtocol":                  reflect.ValueOf(&http.ErrSkipAltProtocol).Elem(),
		"ErrUnexpectedTrailer":                reflect.ValueOf(&http.ErrUnexpectedTrailer).Elem(),
		"ErrUseLastResponse":                  reflect.ValueOf(&http.ErrUseLastResponse).Elem(),
		"ErrWriteAfterFlush":                  reflect.ValueOf(&http.ErrWriteAfterFlush).Elem(),
		"Error":                               reflect.ValueOf(http.Error),
		"FileServer":                          reflect.ValueOf(http.FileServer),
		"Get":                                 reflect.ValueOf(http.Get),
		"Handle":                              reflect.ValueOf(http.Handle),
		"HandleFunc":                          reflect.ValueOf(http.HandleFunc),
		"Head":                                reflect.ValueOf(http.Head),
		"ListenAndServe":                      reflect.ValueOf(http.ListenAndServe),
		"ListenAndServeTLS":                   reflect.ValueOf(http.ListenAndServeTLS),
		"LocalAddrContextKey":                 reflect.ValueOf(&http.LocalAddrContextKey).Elem(),
		"MaxBytesReader":                      reflect.ValueOf(http.MaxBytesReader),
		"MethodConnect":                       reflect.ValueOf(constant.MakeFromLiteral("\"CONNECT\"", token.STRING, 0)),
		"MethodDelete":                        reflect.ValueOf(constant.MakeFromLiteral("\"DELETE\"", token.STRING, 0)),
		"MethodGet":                           reflect.ValueOf(constant.MakeFromLiteral("\"GET\"", token.STRING, 0)),
		"MethodHead":                          reflect.ValueOf(constant.MakeFromLiteral("\"HEAD\"", token.STRING, 0)),
		"MethodOptions":                       reflect.ValueOf(constant.MakeFromLiteral("\"OPTIONS\"", token.STRING, 0)),
		"MethodPatch":                         reflect.ValueOf(constant.MakeFromLiteral("\"PATCH\"", token.STRING, 0)),
		"MethodPost":                          reflect.ValueOf(constant.MakeFromLiteral("\"POST\"", token.STRING, 0)),
		"MethodPut":                           reflect.ValueOf(constant.MakeFromLiteral("\"PUT\"", token.STRING, 0)),
		"MethodTrace":                         reflect.ValueOf(constant.MakeFromLiteral("\"TRACE\"", token.STRING, 0)),
		"NewFileTransport":                    reflect.ValueOf(http.NewFileTransport),
		"NewRequest":                          reflect.ValueOf(http.NewRequest),
		"NewRequestWithContext":               reflect.ValueOf(http.NewRequestWithContext),
		"NewServeMux":                         reflect.ValueOf(http.NewServeMux),
		"NoBody":                              reflect.ValueOf(&http.NoBody).Elem(),
		"NotFound":                            reflect.ValueOf(http.NotFound),
		"NotFoundHandler":                     reflect.ValueOf(http.NotFoundHandler),
		"ParseHTTPVersion":                    reflect.ValueOf(http.ParseHTTPVersion),
		"ParseTime":                           reflect.ValueOf(http.ParseTime),
		"Post":                                reflect.ValueOf(http.Post),
		"PostForm":                            reflect.ValueOf(http.PostForm),
		"ProxyFromEnvironment":                reflect.ValueOf(http.ProxyFromEnvironment),
		"ProxyURL":                            reflect.ValueOf(http.ProxyURL),
		"ReadRequest":                         reflect.ValueOf(http.ReadRequest),
		"ReadResponse":                        reflect.ValueOf(http.ReadResponse),
		"Redirect":                            reflect.ValueOf(http.Redirect),
		"RedirectHandler":                     reflect.ValueOf(http.RedirectHandler),
		"SameSiteDefaultMode":                 reflect.ValueOf(http.SameSiteDefaultMode),
		"SameSiteLaxMode":                     reflect.ValueOf(http.SameSiteLaxMode),
		"SameSiteNoneMode":                    reflect.ValueOf(http.SameSiteNoneMode),
		"SameSiteStrictMode":                  reflect.ValueOf(http.SameSiteStrictMode),
		"Serve":                               reflect.ValueOf(http.Serve),
		"ServeContent":                        reflect.ValueOf(http.ServeContent),
		"ServeFile":                           reflect.ValueOf(http.ServeFile),
		"ServeTLS":                            reflect.ValueOf(http.ServeTLS),
		"ServerContextKey":                    reflect.ValueOf(&http.ServerContextKey).Elem(),
		"SetCookie":                           reflect.ValueOf(http.SetCookie),
		"StateActive":                         reflect.ValueOf(http.StateActive),
		"StateClosed":                         reflect.ValueOf(http.StateClosed),
		"StateHijacked":                       reflect.ValueOf(http.StateHijacked),
		"StateIdle":                           reflect.ValueOf(http.StateIdle),
		"StateNew":                            reflect.ValueOf(http.StateNew),
		"StatusAccepted":                      reflect.ValueOf(constant.MakeFromLiteral("202", token.INT, 0)),
		"StatusAlreadyReported":               reflect.ValueOf(constant.MakeFromLiteral("208", token.INT, 0)),
		"StatusBadGateway":                    reflect.ValueOf(constant.MakeFromLiteral("502", token.INT, 0)),
		"StatusBadRequest":                    reflect.ValueOf(constant.MakeFromLiteral("400", token.INT, 0)),
		"StatusConflict":                      reflect.ValueOf(constant.MakeFromLiteral("409", token.INT, 0)),
		"StatusContinue":                      reflect.ValueOf(constant.MakeFromLiteral("100", token.INT, 0)),
		"StatusCreated":                       reflect.ValueOf(constant.MakeFromLiteral("201", token.INT, 0)),
		"StatusEarlyHints":                    reflect.ValueOf(constant.MakeFromLiteral("103", token.INT, 0)),
		"StatusExpectationFailed":             reflect.ValueOf(constant.MakeFromLiteral("417", token.INT, 0)),
		"StatusFailedDependency":              reflect.ValueOf(constant.MakeFromLiteral("424", token.INT, 0)),
		"StatusForbidden":                     reflect.ValueOf(constant.MakeFromLiteral("403", token.INT, 0)),
		"StatusFound":                         reflect.ValueOf(constant.MakeFromLiteral("302", token.INT, 0)),
		"StatusGatewayTimeout":                reflect.ValueOf(constant.MakeFromLiteral("504", token.INT, 0)),
		"StatusGone":                          reflect.ValueOf(constant.MakeFromLiteral("410", token.INT, 0)),
		"StatusHTTPVersionNotSupported":       reflect.ValueOf(constant.MakeFromLiteral("505", token.INT, 0)),
		"StatusIMUsed":                        reflect.ValueOf(constant.MakeFromLiteral("226", token.INT, 0)),
		"StatusInsufficientStorage":           reflect.ValueOf(constant.MakeFromLiteral("507", token.INT, 0)),
		"StatusInternalServerError":           reflect.ValueOf(constant.MakeFromLiteral("500", token.INT, 0)),
		"StatusLengthRequired":                reflect.ValueOf(constant.MakeFromLiteral("411", token.INT, 0)),
		"StatusLocked":                        reflect.ValueOf(constant.MakeFromLiteral("423", token.INT, 0)),
		"StatusLoopDetected":                  reflect.ValueOf(constant.MakeFromLiteral("508", token.INT, 0)),
		"StatusMethodNotAllowed":              reflect.ValueOf(constant.MakeFromLiteral("405", token.INT, 0)),
		"StatusMisdirectedRequest":            reflect.ValueOf(constant.MakeFromLiteral("421", token.INT, 0)),
		"StatusMovedPermanently":              reflect.ValueOf(constant.MakeFromLiteral("301", token.INT, 0)),
		"StatusMultiStatus":                   reflect.ValueOf(constant.MakeFromLiteral("207", token.INT, 0)),
		"StatusMultipleChoices":               reflect.ValueOf(constant.MakeFromLiteral("300", token.INT, 0)),
		"StatusNetworkAuthenticationRequired": reflect.ValueOf(constant.MakeFromLiteral("511", token.INT, 0)),
		"StatusNoContent":                     reflect.ValueOf(constant.MakeFromLiteral("204", token.INT, 0)),
		"StatusNonAuthoritativeInfo":          reflect.ValueOf(constant.MakeFromLiteral("203", token.INT, 0)),
		"StatusNotAcceptable":                 reflect.ValueOf(constant.MakeFromLiteral("406", token.INT, 0)),
		"StatusNotExtended":                   reflect.ValueOf(constant.MakeFromLiteral("510", token.INT, 0)),
		"StatusNotFound":                      reflect.ValueOf(constant.MakeFromLiteral("404", token.INT, 0)),
		"StatusNotImplemented":                reflect.ValueOf(constant.MakeFromLiteral("501", token.INT, 0)),
		"StatusNotModified":                   reflect.ValueOf(constant.MakeFromLiteral("304", token.INT, 0)),
		"StatusOK":                            reflect.ValueOf(constant.MakeFromLiteral("200", token.INT, 0)),
		"StatusPartialContent":                reflect.ValueOf(constant.MakeFromLiteral("206", token.INT, 0)),
		"StatusPaymentRequired":               reflect.ValueOf(constant.MakeFromLiteral("402", token.INT, 0)),
		"StatusPermanentRedirect":             reflect.ValueOf(constant.MakeFromLiteral("308", token.INT, 0)),
		"StatusPreconditionFailed":            reflect.ValueOf(constant.MakeFromLiteral("412", token.INT, 0)),
		"StatusPreconditionRequired":          reflect.ValueOf(constant.MakeFromLiteral("428", token.INT, 0)),
		"StatusProcessing":                    reflect.ValueOf(constant.MakeFromLiteral("102", token.INT, 0)),
		"StatusProxyAuthRequired":             reflect.ValueOf(constant.MakeFromLiteral("407", token.INT, 0)),
		"StatusRequestEntityTooLarge":         reflect.ValueOf(constant.MakeFromLiteral("413", token.INT, 0)),
		"StatusRequestHeaderFieldsTooLarge":   reflect.ValueOf(constant.MakeFromLiteral("431", token.INT, 0)),
		"StatusRequestTimeout":                reflect.ValueOf(constant.MakeFromLiteral("408", token.INT, 0)),
		"StatusRequestURITooLong":             reflect.ValueOf(constant.MakeFromLiteral("414", token.INT, 0)),
		"StatusRequestedRangeNotSatisfiable":  reflect.ValueOf(constant.MakeFromLiteral("416", token.INT, 0)),
		"StatusResetContent":                  reflect.ValueOf(constant.MakeFromLiteral("205", token.INT, 0)),
		"StatusSeeOther":                      reflect.ValueOf(constant.MakeFromLiteral("303", token.INT, 0)),
		"StatusServiceUnavailable":            reflect.ValueOf(constant.MakeFromLiteral("503", token.INT, 0)),
		"StatusSwitchingProtocols":            reflect.ValueOf(constant.MakeFromLiteral("101", token.INT, 0)),
		"StatusTeapot":                        reflect.ValueOf(constant.MakeFromLiteral("418", token.INT, 0)),
		"StatusTemporaryRedirect":             reflect.ValueOf(constant.MakeFromLiteral("307", token.INT, 0)),
		"StatusText":                          reflect.ValueOf(http.StatusText),
		"StatusTooEarly":                      reflect.ValueOf(constant.MakeFromLiteral("425", token.INT, 0)),
		"StatusTooManyRequests":               reflect.ValueOf(constant.MakeFromLiteral("429", token.INT, 0)),
		"StatusUnauthorized":                  reflect.ValueOf(constant.MakeFromLiteral("401", token.INT, 0)),
		"StatusUnavailableForLegalReasons":    reflect.ValueOf(constant.MakeFromLiteral("451", token.INT, 0)),
		"StatusUnprocessableEntity":           reflect.ValueOf(constant.MakeFromLiteral("422", token.INT, 0)),
		"StatusUnsupportedMediaType":          reflect.ValueOf(constant.MakeFromLiteral("415", token.INT, 0)),
		"StatusUpgradeRequired":               reflect.ValueOf(constant.MakeFromLiteral("426", token.INT, 0)),
		"StatusUseProxy":                      reflect.ValueOf(constant.MakeFromLiteral("305", token.INT, 0)),
		"StatusVariantAlsoNegotiates":         reflect.ValueOf(constant.MakeFromLiteral("506", token.INT, 0)),
		"StripPrefix":                         reflect.ValueOf(http.StripPrefix),
		"TimeFormat":                          reflect.ValueOf(constant.MakeFromLiteral("\"Mon, 02 Jan 2006 15:04:05 GMT\"", token.STRING, 0)),
		"TimeoutHandler":                      reflect.ValueOf(http.TimeoutHandler),
		"TrailerPrefix":                       reflect.ValueOf(constant.MakeFromLiteral("\"Trailer:\"", token.STRING, 0)),

		// type definitions
		"Client":         reflect.ValueOf((*http.Client)(nil)),
		"CloseNotifier":  reflect.ValueOf((*http.CloseNotifier)(nil)),
		"ConnState":      reflect.ValueOf((*http.ConnState)(nil)),
		"Cookie":         reflect.ValueOf((*http.Cookie)(nil)),
		"CookieJar":      reflect.ValueOf((*http.CookieJar)(nil)),
		"Dir":            reflect.ValueOf((*http.Dir)(nil)),
		"File":           reflect.ValueOf((*http.File)(nil)),
		"FileSystem":     reflect.ValueOf((*http.FileSystem)(nil)),
		"Flusher":        reflect.ValueOf((*http.Flusher)(nil)),
		"Handler":        reflect.ValueOf((*http.Handler)(nil)),
		"HandlerFunc":    reflect.ValueOf((*http.HandlerFunc)(nil)),
		"Header":         reflect.ValueOf((*http.Header)(nil)),
		"Hijacker":       reflect.ValueOf((*http.Hijacker)(nil)),
		"ProtocolError":  reflect.ValueOf((*http.ProtocolError)(nil)),
		"PushOptions":    reflect.ValueOf((*http.PushOptions)(nil)),
		"Pusher":         reflect.ValueOf((*http.Pusher)(nil)),
		"Request":        reflect.ValueOf((*http.Request)(nil)),
		"Response":       reflect.ValueOf((*http.Response)(nil)),
		"ResponseWriter": reflect.ValueOf((*http.ResponseWriter)(nil)),
		"RoundTripper":   reflect.ValueOf((*http.RoundTripper)(nil)),
		"SameSite":       reflect.ValueOf((*http.SameSite)(nil)),
		"ServeMux":       reflect.ValueOf((*http.ServeMux)(nil)),
		"Server":         reflect.ValueOf((*http.Server)(nil)),
		"Transport":      reflect.ValueOf((*http.Transport)(nil)),

		// interface wrapper definitions
		"_CloseNotifier":  reflect.ValueOf((*_net_http_CloseNotifier)(nil)),
		"_CookieJar":      reflect.ValueOf((*_net_http_CookieJar)(nil)),
		"_File":           reflect.ValueOf((*_net_http_File)(nil)),
		"_FileSystem":     reflect.ValueOf((*_net_http_FileSystem)(nil)),
		"_Flusher":        reflect.ValueOf((*_net_http_Flusher)(nil)),
		"_Handler":        reflect.ValueOf((*_net_http_Handler)(nil)),
		"_Hijacker":       reflect.ValueOf((*_net_http_Hijacker)(nil)),
		"_Pusher":         reflect.ValueOf((*_net_http_Pusher)(nil)),
		"_ResponseWriter": reflect.ValueOf((*_net_http_ResponseWriter)(nil)),
		"_RoundTripper":   reflect.ValueOf((*_net_http_RoundTripper)(nil)),
	}
}
//<sort>
// _sort_Interface is an interface wrapper for Interface type
type _sort_Interface struct {
	WLen  func() int
	WLess func(i int, j int) bool
	WSwap func(i int, j int)
}

func (W _sort_Interface) Len() int               { return W.WLen() }
func (W _sort_Interface) Less(i int, j int) bool { return W.WLess(i, j) }
func (W _sort_Interface) Swap(i int, j int)      { W.WSwap(i, j) }

// </sort>


// <JSON>
// _encoding_json_Marshaler is an interface wrapper for Marshaler type
type _encoding_json_Marshaler struct {
	WMarshalJSON func() ([]byte, error)
}

func (W _encoding_json_Marshaler) MarshalJSON() ([]byte, error) { return W.WMarshalJSON() }

// _encoding_json_Token is an interface wrapper for Token type
type _encoding_json_Token struct {
}

// _encoding_json_Unmarshaler is an interface wrapper for Unmarshaler type
type _encoding_json_Unmarshaler struct {
	WUnmarshalJSON func(a0 []byte) error
}

func (W _encoding_json_Unmarshaler) UnmarshalJSON(a0 []byte) error { return W.WUnmarshalJSON(a0) }

// </JSON>


// <NET>
// _net_Addr is an interface wrapper for Addr type
type _net_Addr struct {
	WNetwork func() string
	WString  func() string
}

func (W _net_Addr) Network() string { return W.WNetwork() }
func (W _net_Addr) String() string  { return W.WString() }

// _net_Conn is an interface wrapper for Conn type
type _net_Conn struct {
	WClose            func() error
	WLocalAddr        func() net.Addr
	WRead             func(b []byte) (n int, err error)
	WRemoteAddr       func() net.Addr
	WSetDeadline      func(t time.Time) error
	WSetReadDeadline  func(t time.Time) error
	WSetWriteDeadline func(t time.Time) error
	WWrite            func(b []byte) (n int, err error)
}

func (W _net_Conn) Close() error                       { return W.WClose() }
func (W _net_Conn) LocalAddr() net.Addr                { return W.WLocalAddr() }
func (W _net_Conn) Read(b []byte) (n int, err error)   { return W.WRead(b) }
func (W _net_Conn) RemoteAddr() net.Addr               { return W.WRemoteAddr() }
func (W _net_Conn) SetDeadline(t time.Time) error      { return W.WSetDeadline(t) }
func (W _net_Conn) SetReadDeadline(t time.Time) error  { return W.WSetReadDeadline(t) }
func (W _net_Conn) SetWriteDeadline(t time.Time) error { return W.WSetWriteDeadline(t) }
func (W _net_Conn) Write(b []byte) (n int, err error)  { return W.WWrite(b) }

// _net_Error is an interface wrapper for Error type
type _net_Error struct {
	WError     func() string
	WTemporary func() bool
	WTimeout   func() bool
}

func (W _net_Error) Error() string   { return W.WError() }
func (W _net_Error) Temporary() bool { return W.WTemporary() }
func (W _net_Error) Timeout() bool   { return W.WTimeout() }

// _net_Listener is an interface wrapper for Listener type
type _net_Listener struct {
	WAccept func() (net.Conn, error)
	WAddr   func() net.Addr
	WClose  func() error
}

func (W _net_Listener) Accept() (net.Conn, error) { return W.WAccept() }
func (W _net_Listener) Addr() net.Addr            { return W.WAddr() }
func (W _net_Listener) Close() error              { return W.WClose() }

// _net_PacketConn is an interface wrapper for PacketConn type
type _net_PacketConn struct {
	WClose            func() error
	WLocalAddr        func() net.Addr
	WReadFrom         func(p []byte) (n int, addr net.Addr, err error)
	WSetDeadline      func(t time.Time) error
	WSetReadDeadline  func(t time.Time) error
	WSetWriteDeadline func(t time.Time) error
	WWriteTo          func(p []byte, addr net.Addr) (n int, err error)
}

func (W _net_PacketConn) Close() error                                        { return W.WClose() }
func (W _net_PacketConn) LocalAddr() net.Addr                                 { return W.WLocalAddr() }
func (W _net_PacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) { return W.WReadFrom(p) }
func (W _net_PacketConn) SetDeadline(t time.Time) error                       { return W.WSetDeadline(t) }
func (W _net_PacketConn) SetReadDeadline(t time.Time) error                   { return W.WSetReadDeadline(t) }
func (W _net_PacketConn) SetWriteDeadline(t time.Time) error                  { return W.WSetWriteDeadline(t) }
func (W _net_PacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return W.WWriteTo(p, addr)
}

// </NET>

// <HTTP>

// _net_http_CloseNotifier is an interface wrapper for CloseNotifier type
type _net_http_CloseNotifier struct {
	WCloseNotify func() <-chan bool
}

func (W _net_http_CloseNotifier) CloseNotify() <-chan bool { return W.WCloseNotify() }

// _net_http_CookieJar is an interface wrapper for CookieJar type
type _net_http_CookieJar struct {
	WCookies    func(u *url.URL) []*http.Cookie
	WSetCookies func(u *url.URL, cookies []*http.Cookie)
}

func (W _net_http_CookieJar) Cookies(u *url.URL) []*http.Cookie { return W.WCookies(u) }
func (W _net_http_CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	W.WSetCookies(u, cookies)
}

// _net_http_File is an interface wrapper for File type
type _net_http_File struct {
	WClose   func() error
	WRead    func(p []byte) (n int, err error)
	WReaddir func(count int) ([]os.FileInfo, error)
	WSeek    func(offset int64, whence int) (int64, error)
	WStat    func() (os.FileInfo, error)
}

func (W _net_http_File) Close() error                                 { return W.WClose() }
func (W _net_http_File) Read(p []byte) (n int, err error)             { return W.WRead(p) }
func (W _net_http_File) Readdir(count int) ([]os.FileInfo, error)     { return W.WReaddir(count) }
func (W _net_http_File) Seek(offset int64, whence int) (int64, error) { return W.WSeek(offset, whence) }
func (W _net_http_File) Stat() (os.FileInfo, error)                   { return W.WStat() }

// _net_http_FileSystem is an interface wrapper for FileSystem type
type _net_http_FileSystem struct {
	WOpen func(name string) (http.File, error)
}

func (W _net_http_FileSystem) Open(name string) (http.File, error) { return W.WOpen(name) }

// _net_http_Flusher is an interface wrapper for Flusher type
type _net_http_Flusher struct {
	WFlush func()
}

func (W _net_http_Flusher) Flush() { W.WFlush() }

// _net_http_Handler is an interface wrapper for Handler type
type _net_http_Handler struct {
	WServeHTTP func(a0 http.ResponseWriter, a1 *http.Request)
}

func (W _net_http_Handler) ServeHTTP(a0 http.ResponseWriter, a1 *http.Request) { W.WServeHTTP(a0, a1) }

// _net_http_Hijacker is an interface wrapper for Hijacker type
type _net_http_Hijacker struct {
	WHijack func() (net.Conn, *bufio.ReadWriter, error)
}

func (W _net_http_Hijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) { return W.WHijack() }

// _net_http_Pusher is an interface wrapper for Pusher type
type _net_http_Pusher struct {
	WPush func(target string, opts *http.PushOptions) error
}

func (W _net_http_Pusher) Push(target string, opts *http.PushOptions) error {
	return W.WPush(target, opts)
}

// _net_http_ResponseWriter is an interface wrapper for ResponseWriter type
type _net_http_ResponseWriter struct {
	WHeader      func() http.Header
	WWrite       func(a0 []byte) (int, error)
	WWriteHeader func(statusCode int)
}

func (W _net_http_ResponseWriter) Header() http.Header          { return W.WHeader() }
func (W _net_http_ResponseWriter) Write(a0 []byte) (int, error) { return W.WWrite(a0) }
func (W _net_http_ResponseWriter) WriteHeader(statusCode int)   { W.WWriteHeader(statusCode) }

// _net_http_RoundTripper is an interface wrapper for RoundTripper type
type _net_http_RoundTripper struct {
	WRoundTrip func(a0 *http.Request) (*http.Response, error)
}

func (W _net_http_RoundTripper) RoundTrip(a0 *http.Request) (*http.Response, error) {
	return W.WRoundTrip(a0)
}

// </HTTP>