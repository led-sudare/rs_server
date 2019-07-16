package util

import "sync"

type EnumData3DCallback func(x, y, z int, c interface{})
type Data3D interface {
	SetAt(x, y, z int, c interface{})
	GetAt(x, y, z int) interface{}
	Copy() Data3D
	IsInRange(x, y, z int) bool
	ForEach(callback EnumData3DCallback)
	ConcurrentForEach(callback EnumData3DCallback)
	ConcurrentForEachAll(callback EnumData3DCallback)
	Clear()
	Fill(c interface{})
	EditSafe(editor func(data Data3D))
}

type Data3DImpl struct {
	X, Y, Z                   int
	offsetX, offsetY, offsetZ int
	data                      [][][]interface{}
	mutex                     *sync.Mutex
}

func NewData3D(x, y, z, offsetX, offsetY, offsetZ int) Data3D {
	cube := Data3DImpl{
		x, y, z,
		offsetX, offsetY, offsetZ,
		make([][][]interface{}, x),
		new(sync.Mutex)}

	for xx := range cube.data {
		cube.data[xx] = make([][]interface{}, cube.Y)
		for yy := range cube.data[xx] {
			cube.data[xx][yy] = make([]interface{}, cube.Z)
		}
	}
	return &cube
}

func (l *Data3DImpl) IsInRange(x, y, z int) bool {
	switch {
	case 0 > x+l.offsetX:
		fallthrough
	case x+l.offsetX >= l.X:
		fallthrough
	case 0 > y+l.offsetY:
		fallthrough
	case y+l.offsetY >= l.Y:
		fallthrough
	case 0 > z+l.offsetZ:
		fallthrough
	case z+l.offsetZ >= l.Z:
		return false
	}
	return true
}

func (l *Data3DImpl) SetAt(x, y, z int, c interface{}) {
	if l.IsInRange(x, y, z) {
		l.data[x+l.offsetX][y+l.offsetY][z+l.offsetZ] = c
	}
}

func (l *Data3DImpl) GetAt(x, y, z int) interface{} {
	if l.IsInRange(x, y, z) {
		return l.data[x+l.offsetX][y+l.offsetY][z+l.offsetZ]
	} else {
		return nil
	}
}

func (l *Data3DImpl) Copy() Data3D {
	cp := NewData3D(l.X, l.Y, l.Z, l.offsetX, l.offsetY, l.offsetZ)
	l.ConcurrentForEachAll(func(x, y, z int, c interface{}) {
		cp.SetAt(x, y, z, l.GetAt(x, y, z))
	})
	return cp
}

func (l *Data3DImpl) EditSafe(editableBlock func(editable Data3D)) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	editableBlock(l)
}

func (l *Data3DImpl) Clear() {
	ConcurrentEnumXYZ(l.X, l.Y, l.Z, func(x, y, z int) {
		l.SetAt(x, y, z, nil)
	})
}

func (l *Data3DImpl) Fill(c interface{}) {
	ConcurrentEnumXYZ(l.X, l.Y, l.Z, func(x, y, z int) {
		l.SetAt(x, y, z, c)
	})
}

func (l *Data3DImpl) ForEach(callback EnumData3DCallback) {
	EnumXYZ(l.X, l.Y, l.Z, func(x, y, z int) {
		c := l.GetAt(x, y, z)
		if c != nil {
			callback(x, y, z, c)
		}
	})
}
func (l *Data3DImpl) ConcurrentForEach(callback EnumData3DCallback) {
	ConcurrentEnumXYZ(l.X, l.Y, l.Z, func(x, y, z int) {
		c := l.GetAt(x, y, z)
		if c != nil {
			callback(x, y, z, c)
		}
	})
}
func (l *Data3DImpl) ConcurrentForEachAll(callback EnumData3DCallback) {
	ConcurrentEnumXYZ(l.X, l.Y, l.Z, func(x, y, z int) {
		c := l.GetAt(x, y, z)
		callback(x, y, z, c)
	})
}

type EnumXYZCallback func(x, y, z int)

func EnumXYZ(x, y, z int, callback EnumXYZCallback) {
	for xx := 0; xx < x; xx++ {
		for yy := 0; yy < y; yy++ {
			for zz := 0; zz < z; zz++ {
				callback(xx, yy, zz)
			}
		}
	}
}
func ConcurrentEnumXYZ(x, y, z int, callback EnumXYZCallback) {

	ConcurrentEnum(0, x, func(x int) {
		for yy := 0; yy < y; yy++ {
			for zz := 0; zz < z; zz++ {
				callback(x, yy, zz)
			}
		}
	})
}
