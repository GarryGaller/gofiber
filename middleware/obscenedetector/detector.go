package obscenedetector

import (
    "fmt"
    //"regexp"
    "strings"
    "time"
    "unicode/utf8"

    "github.com/dlclark/regexp2"
    "github.com/gofiber/fiber/v2"
)

func ObsceneVocabulary() string {
    var OBSCENE_VOCABULARY = strings.Join(
        []string{`a[\P{L}_]*s[\P{L}_]*s(?:[\P{L}_]*e[\P{L}_]*s)?|f[\P{L}_]*u[\P{L}_]*c[\P{L}_]*k(?:[\P{L}_]*i[\P{L}_]*n[\P{L}_]*`,
            `g)?|ж[\P{L}_]*(?:[ыиiu][\P{L}_]*[дd](?:[\P{L}_]*[уыаyiau]|[\P{L}_]*[оo0][\P{L}_]*[вbv])?|[оo0][`,
            `\P{L}_]*[пnp][\P{L}_]*(?:[аa](?:[\P{L}_]*[хxh])?|[уеыeyiu]|[оo0][\P{L}_]*[йj]))|[дd][\P{L}_]*[е`,
            `e][\P{L}_]*[рpr][\P{L}_]*(?:[ьb][\P{L}_]*)?[мm][\P{L}_]*[оуеаeoya0u](?:[\P{L}_]*[мm])?|[чc][\P{L}_`,
            `]*[мm][\P{L}_]*(?:[оo0]|[ыi][\P{L}_]*[рpr][\P{L}_]*[еиьяeibu])|[сsc][\P{L}_]*[уuy][\P{L}_]*(?:(`,
            `?:[чc][\P{L}_]*)?[кk][\P{L}_]*[ауиiyau](?:[\P{L}_]*[нhn](?:[\P{L}_]*[оo0][\P{L}_]*[йj]|[\P{L}_]*[у`,
            `аыyiau])?)?|[чc][\P{L}_]*(?:(?:[ьb][\P{L}_]*)?(?:[еёяиeiu]|[еиeiu][\P{L}_]*[йj])|[аa][\P{L}_`,
            `]*[рpr][\P{L}_]*[ыауеeyiau]))|[гrg][\P{L}_]*(?:[аоoa0][\P{L}_]*(?:[нhn][\P{L}_]*[дd][\P{L}_]*[а`,
            `оoa0][\P{L}_]*[нhn](?:[\P{L}_]*[ыуyiu])?|[вbv][\P{L}_]*[нhn][\P{L}_]*[оаoa0](?:[\P{L}_]*(?:[мm]`,
            `|[еe][\P{L}_]*[дd](?:[\P{L}_]*[ыуаеeyiau]|[\P{L}_]*[оаoa0][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?)?))?`,
            `)|[нhn][\P{L}_]*(?:[иiu][\P{L}_]*[дd][\P{L}_]*(?:[ыуеаeyiau]|[оo0][\P{L}_]*[йj])|[уyu][сsc](`,
            `?:[\P{L}_]*[аыуyiau]|[\P{L}_]*[оаoa0][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?)?))|(?:[нhn][\P{L}_]*[еe]`,
            `[\P{L}_]*)?(?:(?:[з3z][\P{L}_]*[аa]|[оo0][тt]|[пnp][\P{L}_]*[оo0]|[пnp][\P{L}_]*(?:[еe][\P{L}_]`,
            `*[рpr][\P{L}_]*[еe]|[рpr][\P{L}_]*[оеиeiou0]|[иiu][\P{L}_]*[з3z][\P{L}_]*[дd][\P{L}_]*[оo0])|[н`,
            `hn][\P{L}_]*[аa]|[иiu][\P{L}_]*[з3z]|[дd][\P{L}_]*[оo0]|[вbv][\P{L}_]*[ыi]|[уyu]|[рpr][\P{L}_]*`,
            `[аa][\P{L}_]*[з3z]|[з3z][\P{L}_]*[лl][\P{L}_]*[оo0]|[тt][\P{L}_]*[рpr][\P{L}_]*[оo0]|[уyu])[\P{L}_`,
            `]*)?(?:[вbv][\P{L}_]*[ыi][\P{L}_]*)?(?:[ъьb][\P{L}_]*)?(?:[еёe][\P{L}_]*[бb6](?:(?:[\P{L}_]*[ое`,
            `ёаиуeioyau0])?(?:[\P{L}_]*[нhn](?:[\P{L}_]*[нhn])?[\P{L}_]*[яуаиьiybau]?)?(?:[\P{L}_]*[вbv][`,
            `\P{L}_]*[аa])?(?:(?:[\P{L}_]*(?:[иеeiu]ш[\P{L}_]*[ьb][\P{L}_]*[сsc][\P{L}_]*я|[тt][\P{L}_]*(?:(?:[`,
            `ьb][\P{L}_]*)?[сsc][\P{L}_]*я|[ьb]|[еe][\P{L}_]*[сsc][\P{L}_]*[ьb]|[еe]|[оo0]|[иiu][\P{L}_]*[нh`,
            `n][\P{L}_]*[уыеаeyiau])|(?:щ[\P{L}_]*(?:[иiu][\P{L}_]*[йj]|[аa][\P{L}_]*я|[иеeiu][\P{L}_]*[еe]|`,
            `[еe][\P{L}_]*[гrg][\P{L}_]*[оo0])|ю[\P{L}_]*[тt])(?:[\P{L}_]*[сsc][\P{L}_]*я)?|[еe][\P{L}_]*[мтmt]`,
            `|[кk](?:[\P{L}_]*[иаiau])?|[аa][\P{L}_]*[лl](?:[\P{L}_]*[сsc][\P{L}_]*я)?|[лl][\P{L}_]*(?:[аa][`,
            `\P{L}_]*[нhn]|[оаoa0](?:[\P{L}_]*[мm])?|(?:[иiu][\P{L}_]*)?[сsc][\P{L}_]*[ьяb]|[иiu]|[аa][\P{L}`,
            `_]*[сsc][\P{L}_]*[ьb])|[рpr][\P{L}_]*[ьb]|[сsc][\P{L}_]*[яьb]|[нhn][\P{L}_]*[оo0]|[чc][\P{L}_]*`,
            `(?:[иiu][\P{L}_]*[хxh]|[еe][\P{L}_]*[сsc][\P{L}_]*[тt][\P{L}_]*[ьиibu](?:[\P{L}_]*ю)?)|(?:[тt][`,
            `\P{L}_]*[еe][\P{L}_]*[лl][\P{L}_]*[ьb][\P{L}_]*[сsc][\P{L}_]*[кk][\P{L}_]*|[сsc][\P{L}_]*[тt][\P{L}_]*|[`,
            `лl][\P{L}_]*[иiu][\P{L}_]*[вbv][\P{L}_]*|[чтtc][\P{L}_]*)?(?:[аa][\P{L}_]*я|[оo0][\P{L}_]*[йемejm]`,
            `|[ыi][\P{L}_]*[хйеejxh]|[ыi][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[уyu][\P{L}_]*ю|[иiu][\P{L}_]*[еe]`,
            `|[оo0][\P{L}_]*[мm][\P{L}_]*[уyu]|[иiu][\P{L}_]*[йj]|[еe][\P{L}_]*[вbv]|[иiu][\P{L}_]*[мm](?:[\P{L}_]`,
            `*[иiu])?)|[чтыйилijltcu]))?)|[\P{L}_]*[ыi](?:(?:[\P{L}_]*[вbv][\P{L}_]*[аa]|[\P{L}_]*[нhn`,
            `](?:[\P{L}_]*[нhn])?)(?:(?:[\P{L}_]*(?:[иеeiu]ш[\P{L}_]*[ьb][\P{L}_]*[сsc][\P{L}_]*я|[тt][\P{L}_]*`,
            `(?:[ьb][\P{L}_]*[сsc][\P{L}_]*я|[ьb]|[еe][\P{L}_]*[сsc][\P{L}_]*[ьb]|[еe]|[иiu][\P{L}_]*[нhn][\P{L}_]`,
            `*[уыеаeyiau])|(?:щ[\P{L}_]*(?:[иiu][\P{L}_]*[йj]|[аa][\P{L}_]*я|[иеeiu][\P{L}_]*[еe]|[еe]`,
            `[\P{L}_]*[гrg][\P{L}_]*[оo0])|ю[\P{L}_]*[тt])(?:[\P{L}_]*[сsc][\P{L}_]*я)?|[еe][\P{L}_]*[мтmt]|[лl`,
            `][\P{L}_]*(?:(?:[иiu][\P{L}_]*)?[сsc][\P{L}_]*[ьяb]|[иiu]|[аa][\P{L}_]*[сsc][\P{L}_]*[ьb])|(?:[`,
            `сsc][\P{L}_]*[тt][\P{L}_]*|[лl][\P{L}_]*[иiu][\P{L}_]*[вbv][\P{L}_]*|[чтtc][\P{L}_]*)?(?:[аa][\P{L}_]`,
            `*я|[оo0][\P{L}_]*[йемejm]|[ыi][\P{L}_]*[хйеejxh]|[ыi][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[уyu][`,
            `\P{L}_]*ю|[иiu][\P{L}_]*[еe]|[оo0][\P{L}_]*[мm][\P{L}_]*[уyu]|[иiu][\P{L}_]*[йj]|[еe][\P{L}_]*[вbv`,
            `]|[иiu][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?))))|[рpr][\P{L}_]*[ьb]))|я[\P{L}_]*[бb6](?:[\P{L}_]*[ое`,
            `ёаиуeioyau0])?(?:(?:[\P{L}_]*[нhn](?:[\P{L}_]*[нhn])?[\P{L}_]*[яуаиьiybau]?)?(?:(?:[\P{L}_]*`,
            `(?:[иеeiu]ш[\P{L}_]*[ьb][\P{L}_]*[сsc][\P{L}_]*я|[тt][\P{L}_]*(?:[ьb][\P{L}_]*[сsc][\P{L}_]*я|[ьb]`,
            `|[еe][\P{L}_]*[сsc][\P{L}_]*[ьb]|[еe]|[иiu][\P{L}_]*[нhn][\P{L}_]*[уыеаeyiau])|(?:щ[\P{L}_]*(?:`,
            `[иiu][\P{L}_]*[йj]|[аa][\P{L}_]*я|[иеeiu][\P{L}_]*[еe]|[еe][\P{L}_]*[гrg][\P{L}_]*[оo0])|ю[\P{L}_]`,
            `*[тt])(?:[\P{L}_]*[сsc][\P{L}_]*я)?|[еe][\P{L}_]*[мтmt]|[кk](?:[\P{L}_]*[иаiau])?|[аa][\P{L}_]*`,
            `[лl](?:[\P{L}_]*[сsc][\P{L}_]*я)?|[лl][\P{L}_]*(?:[аa][\P{L}_]*[нhn]|[оаoa0](?:[\P{L}_]*[мm])?|`,
            `(?:[иiu][\P{L}_]*)?[сsc][\P{L}_]*[ьяb]|[иiu])|[рpr][\P{L}_]*[ьb]|[сsc][\P{L}_]*[яьb]|[нhn][\P{L}_]`,
            `*[оo0]|[чc][\P{L}_]*(?:[иiu][\P{L}_]*[хxh]|[еe][\P{L}_]*[сsc][\P{L}_]*[тt][\P{L}_]*[ьиibu](?`,
            `:[\P{L}_]*ю)?)|(?:[тt][\P{L}_]*[еe][\P{L}_]*[лl][\P{L}_]*[ьb][\P{L}_]*[сsc][\P{L}_]*[кk][\P{L}_]*|[сs`,
            `c][\P{L}_]*[тt][\P{L}_]*|[лl][\P{L}_]*[иiu][\P{L}_]*[вbv][\P{L}_]*|[чтtc][\P{L}_]*)?(?:[аa][\P{L}_]*я`,
            `|[оo0][\P{L}_]*[йемejm]|[ыi][\P{L}_]*[хйеejxh]|[ыi][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[уyu][\P{L}`,
            `_]*ю|[иiu][\P{L}_]*[еe]|[оo0][\P{L}_]*[мm][\P{L}_]*[уyu]|[иiu][\P{L}_]*[йj]|[еe][\P{L}_]*[вbv]|`,
            `[иiu][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?)|[чмйилijlmcu]))|(?:[\P{L}_]*[нhn](?:[\P{L}_]*[нhn])?[`,
            `\P{L}_]*[яуаиьiybau]?)))|я[\P{L}_]*[бb6][\P{L}_]*(?:[еёаиуeiyau][\P{L}_]*)?(?:[нhn][\P{L}_]*(?:`,
            `[нhn][\P{L}_]*)?(?:[яуаиьiybau][\P{L}_]*)?)?[тt])|[сsc][\P{L}_]*[ьъb][\P{L}_]*[еяёe][\P{L}_]*[б`,
            `b6][\P{L}_]*(?:[уyu]|(?:[еиёауeiyau](?:[\P{L}_]*[лl](?:[\P{L}_]*[иоаioau0])?|[\P{L}_]*ш[\P{L}_]`,
            `*[ьb]|[\P{L}_]*[тt][\P{L}_]*[еe])?(?:[\P{L}_]*[сsc][\P{L}_]*[ьяb])?))|[еe][\P{L}_]*(?:[бb6][\P{L}_`,
            `]*(?:[уyu][\P{L}_]*[кk][\P{L}_]*[еe][\P{L}_]*[нhn][\P{L}_]*[тt][\P{L}_]*[иiu][\P{L}_]*[йj]|[еe][\P{L}`,
            `_]*[нhn][\P{L}_]*(?:[ьb]|я(?:[\P{L}_]*[мm])?)|[иiu][\P{L}_]*(?:[цc][\P{L}_]*[кk][\P{L}_]*[аa][\P{L}_]`,
            `*я|[чc][\P{L}_]*[еe][\P{L}_]*[сsc][\P{L}_]*[кk][\P{L}_]*[аa][\P{L}_]*я)|[лl][\P{L}_]*[иiu][\P{L}_]`,
            `*щ[\P{L}_]*[еe]|[аa][\P{L}_]*(?:[лl][\P{L}_]*[ьb][\P{L}_]*[нhn][\P{L}_]*[иiu][\P{L}_]*[кk](?:[\P{L}_]`,
            `*[иаiau])?|[тt][\P{L}_]*[оo0][\P{L}_]*[рpr][\P{L}_]*[иiu][\P{L}_]*[йj]|[нhn][\P{L}_]*(?:[тt][\P{L}`,
            `_]*[рpr][\P{L}_]*[оo0][\P{L}_]*[пnp]|[аa][\P{L}_]*[тt][\P{L}_]*[иiu][\P{L}_]*(?:[кk]|[чc][\P{L}_]*`,
            `[еe][\P{L}_]*[сsc][\P{L}_]*[кk][\P{L}_]*[иiu][\P{L}_]*[йj]))))|[дd][\P{L}_]*[рpr][\P{L}_]*[иiu][\P{L}`,
            `_]*[тt])|[нhn][\P{L}_]*[еe][\P{L}_]*[вbv][\P{L}_]*[рpr][\P{L}_]*[оo0][\P{L}_]*[тt][\P{L}_]*ъ[\P{L}_]*`,
            `[еe][\P{L}_]*[бb6][\P{L}_]*[аa][\P{L}_]*[тt][\P{L}_]*[еe][\P{L}_]*[лl][\P{L}_]*[ьb][\P{L}_]*[сsc][\P{L}_`,
            `]*[кk][\P{L}_]*[иiu][\P{L}_]*(?:[ыиiu][\P{L}_]*[йj]|[аa][\P{L}_]*я|[оo0][\P{L}_]*[ейej]|[ыi][\P{L}`,
            `_]*[хxh]|[ыi][\P{L}_]*[еe]|[ыi][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[уyu][\P{L}_]*ю|[оo0][\P{L}_]*[`,
            `мm][\P{L}_]*[уyu])|[уyu][\P{L}_]*(?:[ёеe][\P{L}_]*[бb6][\P{L}_]*(?:[иiu][\P{L}_]*щ[\P{L}_]*[еаea]|`,
            `[аa][\P{L}_]*[нhn](?:[\P{L}_]*[тt][\P{L}_]*[уyu][\P{L}_]*[сsc])?(?:[\P{L}_]*[аоoa0][\P{L}_]*[вмbmv`,
            `]|[\P{L}_]*[ыуеаeyiau])?)|[рpr][\P{L}_]*[оo0][\P{L}_]*[дd](?:[\P{L}_]*[аоoa0][\P{L}_]*[вмbmv]|[`,
            `\P{L}_]*[ыуеаeyiau])?|[бb6][\P{L}_]*[лl][\P{L}_]*ю[\P{L}_]*[дd][\P{L}_]*(?:[оo0][\P{L}_]*[кk]|[кk]`,
            `[\P{L}_]*(?:[аоoa0][\P{L}_]*[вмbmv](?:[\P{L}_]*[иiu])?|[иуеаeiyau])?))|[мm][\P{L}_]*(?:[уyu]`,
            `[\P{L}_]*[дd][\P{L}_]*(?:[оo0][\P{L}_]*[хxh][\P{L}_]*[аa][\P{L}_]*(?:[тt][\P{L}_]*[ьb][\P{L}_]*[сsc][`,
            `\P{L}_]*я|ю[\P{L}_]*[сsc][\P{L}_]*[ьb]|[еe][\P{L}_]*ш[\P{L}_]*[ьb][\P{L}_]*[сsc][\P{L}_]*я)|[аa][\P{L}_]`,
            `*(?:[кk](?:[\P{L}_]*[иаiau]|[оo0][мвbmv])?|[чc][\P{L}_]*(?:[ьb][\P{L}_]*[еёe]|[иiu][\P{L}_]*`,
            `[нhn][\P{L}_]*[уыаyiau]|[кk][\P{L}_]*(?:[аиеуeiyau]|[оo0][\P{L}_]*[йj])))|[еe][\P{L}_]*[нhn]`,
            `[\P{L}_]*[ьb]|[иiu][\P{L}_]*[лl](?:[\P{L}_]*[аеоыeoia0]?))|[аa][\P{L}_]*[нhn][\P{L}_]*[дd][\P{L}_]`,
            `*[уаyau]|[лl][\P{L}_]*(?:[иiu][\P{L}_]*[нhn]|я))|(?:[мm][\P{L}_]*(?:[оo0][\P{L}_]*[з3z][\P{L}_]`,
            `*[гrg]|[уyu][\P{L}_]*[дd])|[дd][\P{L}_]*(?:[оo0][\P{L}_]*[лl][\P{L}_]*[бb6]|[уyu][\P{L}_]*[рpr]`,
            `)|[сsc][\P{L}_]*[кk][\P{L}_]*[оo0][\P{L}_]*[тt]|ж[\P{L}_]*[иiu][\P{L}_]*[дd])[\P{L}_]*[аоoa0][\P{L}_]`,
            `*(?:[хxh][\P{L}_]*[уyu][\P{L}_]*[ийяiju]|[ёеe][\P{L}_]*[бb6](?:[\P{L}_]*[еоeo0][\P{L}_]*[вbv]|[`,
            `\P{L}_]*[ыаia]|[\P{L}_]*[сsc][\P{L}_]*[тt][\P{L}_]*[вbv][\P{L}_]*[оуoy0u](?:[\P{L}_]*[мm])?|[иiu][`,
            `\P{L}_]*[з3z][\P{L}_]*[мm])?)|(?:[нhn][\P{L}_]*[еe][\P{L}_]*|[з3z][\P{L}_]*[аa][\P{L}_]*|[оo0][\P{L}_`,
            `]*[тt][\P{L}_]*|[пnp][\P{L}_]*[оo0][\P{L}_]*|[нhn][\P{L}_]*[аa][\P{L}_]*|[рpr][\P{L}_]*[аa][\P{L}_]*[`,
            `сз3szc][\P{L}_]*)?(?:[пnp][\P{L}_]*[иiu][\P{L}_]*[з3z][\P{L}_]*[дd][\P{L}_]*[ияеeiu]|(?:ъ)?[еёe`,
            `][\P{L}_]*[бb6][\P{L}_]*[аa])[\P{L}_]*(?:(?:[тt][\P{L}_]*[ьb][\P{L}_]*[сsc][\P{L}_]*я|[тt][\P{L}_]*[ь`,
            `b]|[лl][\P{L}_]*[иiu]|[аa][\P{L}_]*[лl]|[лl]|c[\P{L}_]*[ьb]|[иiu][\P{L}_]*[тt]|[иiu]|[тt][\P{L}`,
            `_]*[еe]|[чc][\P{L}_]*[уyu]|ш[\P{L}_]*[ьb])|(?:[йяиiju]|[иеeiu][\P{L}_]*[мm](?:[\P{L}_]*[иiu]`,
            `)?|[йj][\P{L}_]*[сsc][\P{L}_]*(?:[кk][\P{L}_]*(?:[ыиiu][\P{L}_]*[йеej]|[аa][\P{L}_]*я|[оo0][\P{L}_`,
            `]*[еe]|[ыi][\P{L}_]*[хxh]|[ыi][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[уyu][\P{L}_]*ю|[оo0][\P{L}_]*[м`,
            `m][\P{L}_]*[уyu])|[тt][\P{L}_]*[вbv][\P{L}_]*[оуаoya0u](?:[\P{L}_]*[мm])?)))|[пnp][\P{L}_]*[еиы`,
            `eiu][\P{L}_]*[дd][\P{L}_]*[аеэоeoa0][\P{L}_]*[рpr](?:(?:[\P{L}_]*[аa][\P{L}_]*[сз3szc](?:(?:[\P{L}`,
            `_]*[тt])?(?:[\P{L}_]*[ыi]|[\P{L}_]*[оаoa0][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[\P{L}_]*[кk][\P{L}_]*[`,
            `аиiau])?|(?:[\P{L}_]*[ыуаеeyiau]|[\P{L}_]*[оаoa0][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[\P{L}_]*[оo0`,
            `][\P{L}_]*[вbv])))|[\P{L}_]*(?:[ыуаеeyiau]|[оаoa0][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[оo0][\P{L}_`,
            `]*[вbv]|[нhn][\P{L}_]*я))?|[пnp][\P{L}_]*[иiu][\P{L}_]*[з3z][\P{L}_]*(?:[ьb][\P{L}_]*)?[дd][\P{L}_`,
            `]*(?:[ёеe][\P{L}_]*(?:[нhn][\P{L}_]*[ыi][\P{L}_]*ш(?:[\P{L}_]*[ьb])?|[шнжhn](?:[\P{L}_]*[ьb])?)`,
            `|[уyu][\P{L}_]*(?:[йj](?:[\P{L}_]*[тt][\P{L}_]*[еe])?|[нhn](?:[\P{L}_]*[ыi])?)|ю[\P{L}_]*(?:[кk`,
            `](?:[\P{L}_]*(?:[аеуиeiyau]|[оo0][\P{L}_]*[вbv]|[аa][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?))?|[лl]`,
            `(?:[ьиibu]|[еe][\P{L}_]*[йj]|я[\P{L}_]*[хмmxh]))|[еe][\P{L}_]*[цc]|[аоoa0][\P{L}_]*(?:[нhn][`,
            `\P{L}_]*[уyu][\P{L}_]*)?[тt][\P{L}_]*(?:[иiu][\P{L}_]*[йj]|[аa][\P{L}_]*я|[оo0](?:[\P{L}_]*[ейej])`,
            `?|[ыi][\P{L}_]*[ейхejxh]|[ыi][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[уyu][\P{L}_]*ю|[оo0][\P{L}_]*[мm`,
            `][\P{L}_]*[уyu]|[еe][\P{L}_]*[еe]|[ауьеыeyibau])|[аa][\P{L}_]*[нhn][\P{L}_]*[уyu][\P{L}_]*[лl](`,
            `?:[\P{L}_]*[аиiau])?|[ыеуиаeiyau]|[оаoa0][\P{L}_]*(?:[йj]|[хxh][\P{L}_]*[уyu][\P{L}_]*[йj]|[`,
            `еёe][\P{L}_]*[бb6]|(?:[рpr][\P{L}_]*[оo0][\P{L}_]*[тt]|[гrg][\P{L}_]*[оo0][\P{L}_]*[лl][\P{L}_]*[о`,
            `o0][\P{L}_]*[вbv])[\P{L}_]*(?:[ыиiu][\P{L}_]*[йj]|[аa][\P{L}_]*я|[оo0][\P{L}_]*[ейej]|[ыi][\P{L}_]`,
            `*[хxh]|[ыi][\P{L}_]*[еe]|[ыi][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[уyu][\P{L}_]*ю|[оo0][\P{L}_]*[мm`,
            `][\P{L}_]*[уyu])|[бb6][\P{L}_]*(?:[рpr][\P{L}_]*[аa][\P{L}_]*[тt][\P{L}_]*[иiu][\P{L}_]*я|[оo0][\P{L}`,
            `_]*[лl](?:[\P{L}_]*[аыуyiau])?)))|[пnp][\P{L}_]*(?:[аa][\P{L}_]*[дd][\P{L}_]*[лl][\P{L}_]*[аоыo`,
            `ia0]|[оаoa0][\P{L}_]*[сsc][\P{L}_]*[кk][\P{L}_]*[уyu][\P{L}_]*[дd][\P{L}_]*(?:[ыуаеeyiau]|[оаoa`,
            `0][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?)|[иеeiu][\P{L}_]*[дd][\P{L}_]*(?:[иiu][\P{L}_]*[кk]|[рpr][\P{L}`,
            `_]*[иiu][\P{L}_]*[лl](?:[\P{L}_]*[лl])?)(?:[\P{L}_]*[оаoa0][\P{L}_]*[мвbmv]|[\P{L}_]*[иуеоыаeio`,
            `yau0])?|[рpr][\P{L}_]*[оo0][\P{L}_]*[бb6][\P{L}_]*[лl][\P{L}_]*я[\P{L}_]*[дd][\P{L}_]*[оo0][\P{L}_]*[`,
            `мm])|(?:[з3z][\P{L}_]*[аa][\P{L}_]*|[оo0][\P{L}_]*[тt][\P{L}_]*|[нhn][\P{L}_]*[аa][\P{L}_]*)?[сsc]`,
            `[\P{L}_]*[рpr][\P{L}_]*(?:[аa][\P{L}_]*[тt][\P{L}_]*[ьb]|[аa][\P{L}_]*[лl](?:[\P{L}_]*[иiu])?|[eуи`,
            `iyu])|[сsc][\P{L}_]*[рpr][\P{L}_]*[аa][\P{L}_]*(?:[кk][\P{L}_]*(?:[аеиуeiyau]|[оo0][\P{L}_]*[йj`,
            `])|[нhn](?:[\P{L}_]*[нhn])?(?:[ьb]|(?:[\P{L}_]*[ыi][\P{L}_]*[йеej]|[\P{L}_]*[аa][\P{L}_]*я|[\P{L}_`,
            `]*[оo0][\P{L}_]*[еe]))|[лl][\P{L}_]*[ьb][\P{L}_]*[нhn][\P{L}_]*[иiu][\P{L}_]*[кk](?:[\P{L}_]*[иiu]`,
            `|[\P{L}_]*[оаoa0][\P{L}_]*[мm])?)|(?:[з3z][\P{L}_]*[аa][\P{L}_]*)?[тt][\P{L}_]*[рpr][\P{L}_]*[аa][`,
            `\P{L}_]*[хxh][\P{L}_]*(?:[нhn][\P{L}_]*(?:[уyu](?:[\P{L}_]*[тt][\P{L}_]*[ьb](?:[\P{L}_]*[сsc][\P{L}_]`,
            `*я)?|[\P{L}_]*[сsc][\P{L}_]*[ьb]|[\P{L}_]*[лl](?:[\P{L}_]*[аиiau])?)?|[еиeiu][\P{L}_]*ш[\P{L}_]*[ь`,
            `b][\P{L}_]*[сsc][\P{L}_]*я)|[аa][\P{L}_]*(?:[лl](?:[\P{L}_]*[аоиioau0])?|[тt][\P{L}_]*[ьb](?:[\P{L}_]`,
            `*[сsc][\P{L}_]*я)?|[нhn][\P{L}_]*(?:[нhn][\P{L}_]*)?(?:[ыиiu][\P{L}_]*[йj]|[аa][\P{L}_]*я|[о`,
            `o0][\P{L}_]*[йеej]|[ыi][\P{L}_]*[хxh]|[ыi][\P{L}_]*[еe]|[ыi][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[у`,
            `yu][\P{L}_]*ю|[оo0][\P{L}_]*[мm][\P{L}_]*[уyu])))|(?:[нhn][\P{L}_]*[иеeiu][\P{L}_]*|[пnp][\P{L}_]*`,
            `[оo0][\P{L}_]*|[нhn][\P{L}_]*[аa][\P{L}_]*|[оаoa0][\P{L}_]*(?:[тt][\P{L}_]*)?|[дd][\P{L}_]*[аоoa0]`,
            `[\P{L}_]*|[з3z][\P{L}_]*[аa][\P{L}_]*)?(?:(?:[фf][\P{L}_]*[иiu][\P{L}_]*[гrg]|[хxh][\P{L}_]*(?:[еи`,
            `eiu][\P{L}_]*(?:[йj][\P{L}_]*)?[рpr]|[рpr][\P{L}_]*[еe][\P{L}_]*[нhn]|[уyu](?:[\P{L}_]*[йj])?))`,
            `(?:[\P{L}_]*[еоёeo0][\P{L}_]*[вbv](?:[\P{L}_]*[аa][\P{L}_]*ю[\P{L}_]*щ|[\P{L}_]*ш)?)?(?:[\P{L}_]*[аие`,
            `eiau][\P{L}_]*[лнlhn])?(?:[нhn])?(?:[\P{L}_]*(?:[иаоёяыеeioau0][юяиевмйbeijmvu]|я[\P{L}_]`,
            `*(?:[мm](?:[\P{L}_]*[иiu])?|[рpr][\P{L}_]*(?:ю|[иiu][\P{L}_]*(?:[тt](?:[\P{L}_]*[ьеeb][\P{L}_]*`,
            `[сsc][\P{L}_]*[яьb])?|[лl](?:[\P{L}_]*[иоаioau0])?))|[чc][\P{L}_]*(?:[аиiau][\P{L}_]*[тt](?:`,
            `[\P{L}_]*[сsc][\P{L}_]*я)|[иiu][\P{L}_]*[лl](?:[\P{L}_]*[иоаioau0])?)|[чc](?:[\P{L}_]*[ьb])?)|[`,
            `еe][\P{L}_]*(?:[тt][\P{L}_]*(?:[оo0][\P{L}_]*[йj]|[аьуybau])|[еe][\P{L}_]*(?:[тt][\P{L}_]*[еe]|`,
            `ш[\P{L}_]*[ьb]))|[аыоуяюйиijoyau0]|[лl][\P{L}_]*[иоiou0]|[чc][\P{L}_]*[уyu])))|(?:[фf][\P{L}`,
            `_]*[иiu][\P{L}_]*[гrg]|[хxh][\P{L}_]*(?:[еиeiu][\P{L}_]*(?:[йj][\P{L}_]*)?[рpr]|[рpr][\P{L}_]*[`,
            `еe][\P{L}_]*[нhn]|[уyu][\P{L}_]*[йj]))|[хxh][\P{L}_]*[уyu][\P{L}_]*(?:[еёиeiu][\P{L}_]*(?:[сsc]`,
            `[\P{L}_]*[оo0][\P{L}_]*[сsc]|[пnp][\P{L}_]*[лl][\P{L}_]*[еe][\P{L}_]*[тt]|[нhn][\P{L}_]*[ыi][\P{L}_]*`,
            `ш)(?:[\P{L}_]*[аыуyiau]|[\P{L}_]*[оаoa0][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?|[нhn][\P{L}_]*(?:[ыиiu`,
            `][\P{L}_]*[йj]|[аa][\P{L}_]*я|[оo0][\P{L}_]*[йеej]|[ыi][\P{L}_]*[хxh]|[ыi][\P{L}_]*[еe]|[ыi][\P{L}`,
            `_]*[мm](?:[\P{L}_]*[иiu])?|[уyu][\P{L}_]*ю|[оo0][\P{L}_]*[мm][\P{L}_]*[уyu]))?|[дd][\P{L}_]*[оo`,
            `0][\P{L}_]*ё[\P{L}_]*[бb6][\P{L}_]*[иiu][\P{L}_]*[нhn][\P{L}_]*(?:[оo0][\P{L}_]*[йj]|[аеыуeyiau]))`,
            `|[бb6][\P{L}_]*[лl][\P{L}_]*я(?:[\P{L}_]*[дтdt][\P{L}_]*(?:[ьb]|[иiu]|[кk][\P{L}_]*[иiu]|[сsc][`,
            `\P{L}_]*[тt][\P{L}_]*[вbv][\P{L}_]*[оo0]|[сsc][\P{L}_]*[кk][\P{L}_]*(?:[оo0][\P{L}_]*[ейej]|[иiu][`,
            `\P{L}_]*[еe]|[аa][\P{L}_]*я|[иiu][\P{L}_]*[йj]|[оo0][\P{L}_]*[гrg][\P{L}_]*[оo0])))?|[вbv][\P{L}_]`,
            `*[ыi][\P{L}_]*[бb6][\P{L}_]*[лl][\P{L}_]*я[\P{L}_]*[дd][\P{L}_]*(?:[оo0][\P{L}_]*[кk]|[кk][\P{L}_]*(?`,
            `:[иуаеeiyau]|[аa][\P{L}_]*[мm](?:[\P{L}_]*[иiu])?))|(?:[з3z][\P{L}_]*[аоoa0][\P{L}_]*)(?:[пn`,
            `p][\P{L}_]*[аоoa0][\P{L}_]*[дd][\P{L}_]*[лl][\P{L}_]*[оыаoia0]|[лl][\P{L}_]*[уyu][\P{L}_]*[пnp][\P{L}`,
            `_]*(?:[оo0][\P{L}_]*[йj]|[аеыуeyiau]))|ш[\P{L}_]*[лl][\P{L}_]*ю[\P{L}_]*[хxh][\P{L}_]*(?:[ауеиe`,
            `iyau]|[оo0][\P{L}_]*[йj])|[аa][\P{L}_]*[нhn][\P{L}_]*[уyu][\P{L}_]*[сsc](?:[\P{L}_]*[еаыуeyiau]`,
            `|[\P{L}_]*[оo0][\P{L}_]*[мm])?|(?:\p{L}*(?:[хxh](?:[рpr][еe][нhn]|[уyu][иiu])|[пnp][еиeiu`,
            `](?:[з3z][дd]|[дd](?:[еаоeoa0][рpr]|[рpr]))|[бb6][лl]я[дd]|[оo0][хxh][уyu][еe]|[`,
            `мm][уyu][дd][еоиаeioau0]|[дd][еe][рpr][ьb]|[гrg][аоoa0][вbv][нhn]|[уyu][еёe][бb6`,
            `])|[хxh][\P{L}_]*(?:[рpr][\P{L}_]*[еe][\P{L}_]*[нhn]|[уyu][\P{L}_]*[йиеяeiju])|[пnp][\P{L}_]*[е`,
            `иeiu][\P{L}_]*(?:[з3z][\P{L}_]*[дd]|[дd][\P{L}_]*(?:[еаоeoa0][\P{L}_]*[рpr]|[рpr]))|[бb6][\P{L}`,
            `_]*[лl][\P{L}_]*я[\P{L}_]*[дd]|[оo0][\P{L}_]*[хxh][\P{L}_]*[уyu][\P{L}_]*[еe]|[мm][\P{L}_]*[уyu][\P{L}_]`,
            `*[дd][\P{L}_]*[еоиаeioau0]|[дd][\P{L}_]*[еe][\P{L}_]*[рpr][\P{L}_]*[ьb]|[гrg][\P{L}_]*[аоoa0`,
            `][\P{L}_]*[вbv][\P{L}_]*[нhn]|[уyu][\P{L}_]*[еёe][\P{L}_]*[бb6]|[ёеe][бb6])\p{L}+`}, "")

    OBSCENE_VOCABULARY = strings.Join([]string{`(?:\b|(?<=_))(?:`, OBSCENE_VOCABULARY, `)(?:\b|(?=_))`}, "")
    return OBSCENE_VOCABULARY
}

var OBSCENE = regexp2.MustCompile(ObsceneVocabulary(), regexp2.RE2|regexp2.IgnoreCase)

func FindAllString(re *regexp2.Regexp, s string, timeout ...time.Duration) []string {
    var matches []string
    if len(timeout) != 0 {
        re.MatchTimeout = timeout[0]
    }
    m, _ := re.FindStringMatch(s)
    for m != nil {
        matches = append(matches, m.String())
        m, _ = re.FindNextMatch(m)
    }
    return matches
}

func FindString(re *regexp2.Regexp, s string, timeout ...time.Duration) string {
    var result string
    if len(timeout) != 0 {
        re.MatchTimeout = timeout[0]
    }
    m, _ := re.FindStringMatch(s)
    if m != nil {
        result = m.String()
    }

    return result
}

func MatchString(re *regexp2.Regexp, s string, timeout ...time.Duration) bool {
    if len(timeout) != 0 {
        re.MatchTimeout = timeout[0]
    }
    m, _ := re.MatchString(s)
    return m
}

func IsObscene(s string, timeout time.Duration) string {
    return FindString(OBSCENE, s, timeout)
}

func IsObsceneList(s string, timeout time.Duration) []string {
    return FindAllString(OBSCENE, s, timeout)
}

var WhiteList = []*regexp2.Regexp{
    regexp2.MustCompile(`фиг|fig|хрен|her`,
        regexp2.RE2|regexp2.IgnoreCase,
    ),
}

type Config struct {
    Targets      []string
    MaxLen       int
    WhiteList    []*regexp2.Regexp
    RegexTimeout time.Duration
    Next         func(c *fiber.Ctx) bool
    Response     func(c *fiber.Ctx) error
    Local        bool
}

func GetIPs(c *fiber.Ctx) []string {
    ips := c.IPs()
    if len(ips) == 0 {
        ips = append(ips, c.IP())
    }
    return ips
}

func NextIfLocal(c *fiber.Ctx) bool {
    return GetIPs(c)[0] == "127.0.0.1"
}

func Next(c *fiber.Ctx) bool {
    return false
}

func NextIfNotRu(c *fiber.Ctx) bool {
    lang := c.Locals("lang").(string)
    langFrom := strings.Split(lang, "-")
    // пропускаем если исходный язык русский
    if len(langFrom) > 1 {
        if langFrom[0] == "ru" {
            return false
        }
    }

    return true
}

func Response(c *fiber.Ctx) error {
    message := c.Locals("message").(string)

    return c.Status(400).JSON(fiber.Map{
        "status":  fiber.StatusBadRequest,
        "message": message,
    })
}

var ConfigDefault = Config{
    Targets:      make([]string, 0),
    MaxLen:       1000,
    WhiteList:    make([]*regexp2.Regexp, 0),
    RegexTimeout: 50 * time.Millisecond,
    Next:         Next,
    Response:     Response,
}

func configDefault(config ...Config) Config {
    // Return default config if nothing provided
    if len(config) < 1 {
        return ConfigDefault
    }

    // Override default config
    cfg := config[0]

    // Set default values
    if cfg.MaxLen == 0 {
        cfg.MaxLen = ConfigDefault.MaxLen
    }

    if cfg.WhiteList == nil {
        cfg.WhiteList = ConfigDefault.WhiteList
    }

    if cfg.RegexTimeout == 0 {
        cfg.RegexTimeout = ConfigDefault.RegexTimeout
    }

    if cfg.Next == nil {
        cfg.Next = ConfigDefault.Next
    }    
    
    if cfg.Local {
        cfg.Next = NextIfLocal
    }

    if cfg.Response == nil {
        cfg.Response = ConfigDefault.Response
    }

    return cfg
}

func New(config ...Config) fiber.Handler {

    cfg := configDefault(config...)
    // Return new handler
    return func(c *fiber.Ctx) error {
        // Don't execute middleware if Next returns true
        if cfg.Next != nil && cfg.Next(c) {
            return c.Next()
        }

        for _, target := range cfg.Targets {
            value := c.Locals(target).(string)
            runesCount := utf8.RuneCountInString(value)
            if cfg.MaxLen != -1 && runesCount > cfg.MaxLen {
                value = string([]rune(value)[:cfg.MaxLen])
            }
            foundMatches := IsObsceneList(value, cfg.RegexTimeout)
            lenMatches := len(foundMatches)
            if lenMatches != 0 {
                //cleanedMatches := make([]string,0)
                //for _, found := range foundMatches {
                //    for _, re := range cfg.WhiteList {
                //        if !MatchString(re, found, cfg.RegexTimeout) {
                //            cleanedMatches = append(cleanedMatches, found)
                //        }
                //    }
                //}
                //if len(cleanedMatches) != 0 {
                if len(cfg.WhiteList) != 0 {
                    for i := 0; i < len(foundMatches); i++ {
                        for _, re := range cfg.WhiteList {
                            if MatchString(re, foundMatches[i], cfg.RegexTimeout) {
                                foundMatches = append(foundMatches[:i], foundMatches[i+1:]...)
                                i--
                            }
                        }
                    }
                }
                if len(foundMatches) != 0 {
                    message := fmt.Sprintf(
                        "The text contains obscene words: %s...",
                        strings.Join(foundMatches, ","))
                    c.Locals("message", message)
                    return cfg.Response(c)
                }
            }
        }
        return c.Next()
    }
}
