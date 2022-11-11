/* eslint-disable */
// prettier-disable
// prettier-ignore
// Code generated by api2. DO NOT EDIT.
import {route} from "./utils"

export const api = {
example: {
	IEchoService: {
			Hello: route<example.HelloRequest, example.HelloResponse>(
				"POST", "/hello",
				{"query":["key"]},
				{"header":["session"]}),
			Echo: route<example.EchoRequest, example.EchoResponse>(
				"POST", "/echo",
				{"header":["session"],"json":["text","bar","code","dir","items","maps"]},
				{"json":["text"]}),
			Since: route<example.SinceRequest, example.SinceResponse>(
				"POST", "/since",
				{"header":["session"]},
				{}),
			Stream: route<example.StreamRequest, example.StreamResponse>(
				"PUT", "/stream",
				{"header":["session"]},
				{}),
			Redirect: route<example.RedirectRequest, example.RedirectResponse>(
				"GET", "/redirect",
				{"query":["id"]},
				{"header":["Location"]}),
	},
},
}
export const OpCodeEnum  = {
    "Op_Add": ,
    "Op_Read": ,
    "Op_Write": ,
} as const
export const DirectionEnum  = {
    "East": 1,
    "North": 0,
    "South": 2,
    "West": 3,
} as const

export declare namespace example {

export type HelloRequest = {
	key?: string
}


export type HelloResponse = {
	session?: string
}


export type EchoRequest = {
	session?: string
	text: string
	bar: number
	code: example.OpCode
	dir: example.Direction
	items: Array<example.CustomType2> | null
	maps: Record<string, example.Direction> |  null
}


export type OpCode = typeof OpCodeEnum[keyof typeof OpCodeEnum]

export type Direction = typeof DirectionEnum[keyof typeof DirectionEnum]

export type CustomType2 =  example.UserSettings & {
}


export type UserSettings = Record<string, any> |  null
// EchoResponse.
export type EchoResponse = {
	text: string // field comment.
}


export type SinceRequest = {
	session?: string
}


export type SinceResponse = {
}


export type StreamRequest = {
	session?: string
}


export type StreamResponse = {
}


export type RedirectRequest = {
	id?: string
}


export type RedirectResponse = {
	Location?: string
}


export type CustomType =  example.UserSettings & {
}

}
