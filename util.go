package hackberry

import (
    "strconv"
)

func parseBool(s string) bool{
    if b, err := strconv.ParseBool(s); err == nil{
        return b
    }else{
        panic(&ParseError{"Can't parse bool value from string [" + s + "]."})
    }
}

func parseInt(s string) int64{
    if i, err := strconv.ParseInt(s, 10, 64); err == nil{
        return i
    }else{
        panic(&ParseError{"Can't parse int value from string [" + s + "]."})
    }
}

func parseUint(s string) uint64{
    if u, err := strconv.ParseUint(s, 10, 64); err == nil{
        return u
    }else{
        panic(&ParseError{"Can't parse uint value from string [" + s + "]."})
    }
}

func parseFloat(s string) float64{
    if f, err := strconv.ParseFloat(s, 64); err == nil{
        return f
    }else{
        panic(&ParseError{"Can't parse float value from string [" + s + "]."})
    }
}
